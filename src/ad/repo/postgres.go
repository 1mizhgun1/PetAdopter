package repo

import (
	"context"
	goerrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/ad"
)

const (
	getAd = `
SELECT
	Ad.id, Ad.owner_id, Ad.status,
	Ad.photo_url, Ad.title, Ad.description, Ad.price, Ad.animal_id, Ad.breed_id, Ad.contacts,
	Ad.created_at, Ad.updated_at,
	MyUser.username, Animal.name as animal_name, Breed.name as breed_name, Locality.name as locality_name
FROM
    Ad, MyUser, Animal, Breed, Locality
WHERE
    Ad.id=$1
	AND Ad.owner_id = MyUser.id
	AND Ad.animal_id = Animal.id
	AND Ad.breed_id = Breed.id
	AND MyUser.locality_id = Locality.id;
`

	createAd = "INSERT INTO Ad(id, owner_id, status, photo_url, title, description, price, animal_id, breed_id, contacts, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);"
	deleteAd = "DELETE FROM Ad WHERE id=$1;"
)

type AdPostgres struct {
	db pgxtype.Querier
}

func NewAdPostgres(db pgxtype.Querier) *AdPostgres {
	return &AdPostgres{db: db}
}

func (repo *AdPostgres) SearchAds(ctx context.Context, params ad.SearchParams) ([]ad.RespAd, error) {
	query := `
SELECT
	Ad.id, Ad.owner_id, Ad.status,
	Ad.photo_url, Ad.title, Ad.description, Ad.price, Ad.animal_id, Ad.breed_id, Ad.contacts,
	Ad.created_at, Ad.updated_at,
	MyUser.username, Animal.name as animal_name, Breed.name as breed_name, Locality.name as locality_name
FROM
    Ad, MyUser, Animal, Breed, Locality
WHERE
	Ad.owner_id = MyUser.id
	AND Ad.animal_id = Animal.id
	AND Ad.breed_id = Breed.id
	AND MyUser.locality_id = Locality.id
`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.owner_id=$%d", argIndex))
		args = append(args, *params.OwnerID)
		argIndex++
	}

	if params.AnimalID != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.animal_id=$%d", argIndex))
		args = append(args, *params.AnimalID)
		argIndex++
	}

	if params.BreedID != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.breed_id=$%d", argIndex))
		args = append(args, *params.BreedID)
		argIndex++
	}

	if params.MinPrice != nil && params.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.price BETWEEN $%d AND $%d", argIndex, argIndex+1))
		args = append(args, *params.MinPrice)
		args = append(args, *params.MaxPrice)
		argIndex += 2
	} else if params.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.price >= $%d", argIndex))
		args = append(args, *params.MinPrice)
		argIndex++
	} else if params.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("Ad.price <= $%d", argIndex))
		args = append(args, *params.MaxPrice)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d;", argIndex, argIndex+1)

	args = append(args, params.Limit)
	args = append(args, params.Offset)

	result := make([]ad.RespAd, 0)

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return result, errors.Wrap(err, "failed to get ads from postgres")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			row      ad.Ad
			rowExtra ad.AdInfo
		)
		if err = rows.Scan(&row.ID, &row.OwnerID, &row.Status, &row.PhotoURL, &row.Title, &row.Description, &row.Price, &row.AnimalID, &row.BreedID, &row.Contacts, &row.CreatedAt, &row.UpdatedAt, &rowExtra.Username, &rowExtra.AnimalName, &rowExtra.BreedName, &rowExtra.LocalityName); err != nil {
			return result, errors.Wrap(err, "failed to parse ad")
		}
		result = append(result, ad.RespAd{Info: row, ExtraInfo: rowExtra})
	}

	return result, nil
}

func (repo *AdPostgres) GetAd(ctx context.Context, id uuid.UUID) (ad.RespAd, error) {
	result := ad.Ad{}
	resultExtra := ad.AdInfo{}
	if err := repo.db.QueryRow(ctx, getAd, id).Scan(&result.ID, &result.OwnerID, &result.Status, &result.PhotoURL, &result.Title, &result.Description, &result.Price, &result.AnimalID, &result.BreedID, &result.Contacts, &result.CreatedAt, &result.UpdatedAt, &resultExtra.Username, &resultExtra.AnimalName, &resultExtra.BreedName, &resultExtra.LocalityName); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return ad.RespAd{}, ad.ErrAdNotFound
		}
		return ad.RespAd{}, errors.Wrap(err, "failed to get ad from postgres")
	}

	return ad.RespAd{Info: result, ExtraInfo: resultExtra}, nil
}

func (repo *AdPostgres) CreateAd(ctx context.Context, adData ad.Ad) error {
	if _, err := repo.db.Exec(ctx, createAd, adData.ID, adData.OwnerID, adData.Status, adData.PhotoURL, adData.Title, adData.Description, adData.Price, adData.AnimalID, adData.BreedID, adData.Contacts, adData.CreatedAt, adData.UpdatedAt); err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return ad.ErrInvalidForeignKey
		}
		return errors.Wrap(err, "failed to create ad in postgres")
	}

	return nil
}

func (repo *AdPostgres) UpdateAd(ctx context.Context, id uuid.UUID, form ad.UpdateForm, now time.Time) error {
	query := "UPDATE Ad SET "
	var conditions []string
	var args []interface{}
	argIndex := 1

	if form.PhotoURL != nil {
		conditions = append(conditions, fmt.Sprintf("photo_url=$%d", argIndex))
		args = append(args, *form.PhotoURL)
		argIndex++
	}

	if form.Title != nil {
		conditions = append(conditions, fmt.Sprintf("title=$%d", argIndex))
		args = append(args, *form.Title)
		argIndex++
	}

	if form.Description != nil {
		conditions = append(conditions, fmt.Sprintf("description=$%d", argIndex))
		args = append(args, *form.Description)
		argIndex++
	}

	if form.Contacts != nil {
		conditions = append(conditions, fmt.Sprintf("contacts=$%d", argIndex))
		args = append(args, *form.Contacts)
		argIndex++
	}

	if form.Price != nil {
		conditions = append(conditions, fmt.Sprintf("price=$%d", argIndex))
		args = append(args, *form.Price)
		argIndex++
	}

	if form.AnimalID != nil {
		conditions = append(conditions, fmt.Sprintf("animal_id=$%d", argIndex))
		args = append(args, *form.AnimalID)
		argIndex++
	}

	if form.BreedID != nil {
		conditions = append(conditions, fmt.Sprintf("breed_id=$%d", argIndex))
		args = append(args, *form.BreedID)
		argIndex++
	}

	if form.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status=$%d", argIndex))
		args = append(args, *form.Status)
		argIndex++
	}

	conditions = append(conditions, fmt.Sprintf("updated_at=$%d", argIndex))
	args = append(args, now)
	argIndex++

	if len(conditions) == 0 {
		return nil
	}

	query += strings.Join(conditions, ", ") + fmt.Sprintf(" WHERE id=$%d;", argIndex)

	args = append(args, id)

	if _, err := repo.db.Exec(ctx, query, args...); err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return ad.ErrInvalidForeignKey
		}
		return errors.Wrap(err, "failed to update ad in postgres")
	}

	return nil
}

func (repo *AdPostgres) DeleteAd(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, deleteAd, id); err != nil {
		return errors.Wrap(err, "failed to delete ad from postgres")
	}

	return nil
}
