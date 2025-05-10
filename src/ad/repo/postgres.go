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
	MyUser.username,
	Animal.name AS animal_name,
	Breed.name AS breed_name,
	COALESCE(Locality.name, '') AS locality_name
FROM Ad
JOIN MyUser ON Ad.owner_id = MyUser.id
JOIN Animal ON Ad.animal_id = Animal.id
JOIN Breed ON Ad.breed_id = Breed.id
LEFT JOIN Locality ON MyUser.locality_id = Locality.id
WHERE Ad.id = $1;
`

	createAd = "INSERT INTO Ad(id, owner_id, status, photo_url, title, description, price, animal_id, breed_id, contacts, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);"
	deleteAd = "DELETE FROM Ad WHERE id=$1;"

	getHistory  = "SELECT user_id, animal_id, breed_id, min_price, max_price, radius, created_at FROM History WHERE user_id = $1;"
	saveHistory = `
INSERT INTO
	History(user_id, animal_id, breed_id, min_price, max_price, radius, created_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7)
ON
	CONFLICT (user_id) DO
UPDATE
    SET animal_id = $2, breed_id = $3, min_price = $4, max_price = $5, radius = $6, created_at = $7
`
)

type AdPostgres struct {
	db pgxtype.Querier
}

func NewAdPostgres(db pgxtype.Querier) *AdPostgres {
	return &AdPostgres{db: db}
}

func (repo *AdPostgres) SearchAds(ctx context.Context, params ad.SearchParams, extra ad.SearchExtra) ([]ad.RespAd, error) {
	query := `
SELECT
	Ad.id, Ad.owner_id, Ad.status,
	Ad.photo_url, Ad.title, Ad.description, Ad.price, Ad.animal_id, Ad.breed_id, Ad.contacts,
	Ad.created_at, Ad.updated_at,
	MyUser.username,
	Animal.name AS animal_name,
	Breed.name AS breed_name,
	COALESCE(Locality.name, '') AS locality_name,
	COALESCE(Locality.latitude, NULL) AS locality_latitude,
	COALESCE(Locality.longitude, NULL) AS locality_longitude
FROM Ad
JOIN MyUser ON Ad.owner_id = MyUser.id
JOIN Animal ON Ad.animal_id = Animal.id
JOIN Breed ON Ad.breed_id = Breed.id
LEFT JOIN Locality ON MyUser.locality_id = Locality.id
`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if !params.AllStatuses {
		conditions = append(conditions, "Ad.status = 'A'")
	}

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
		args = append(args, *params.MinPrice, *params.MaxPrice)
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

	if params.Radius != nil {
		conditions = append(conditions, fmt.Sprintf("(Locality.latitude IS NULL OR Locality.longitude IS NULL OR haversine_distance(Locality.latitude, Locality.longitude, $%d, $%d) <= $%d)", argIndex, argIndex+1, argIndex+2))
		args = append(args, extra.Latitude, extra.Longitude, *params.Radius)
		argIndex += 3
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if extra.Best == nil {
		query += " ORDER BY Ad.updated_at DESC "
	} else {
		scoreParts := make([]string, 0)

		if extra.Best.AnimalID == nil {
			scoreParts = append(scoreParts, "1")
		} else {
			scoreParts = append(scoreParts, fmt.Sprintf("CASE WHEN Ad.animal_id = $%d THEN 1 ELSE 0 END", argIndex))
			args = append(args, *extra.Best.AnimalID)
			argIndex++
		}

		if extra.Best.BreedID == nil {
			scoreParts = append(scoreParts, "1")
		} else {
			scoreParts = append(scoreParts, fmt.Sprintf("CASE WHEN Ad.breed_id = $%d THEN 1 ELSE 0 END", argIndex))
			args = append(args, *extra.Best.BreedID)
			argIndex++
		}

		if extra.Best.MinPrice == nil {
			scoreParts = append(scoreParts, "1")
		} else {
			scoreParts = append(scoreParts, fmt.Sprintf("CASE WHEN Ad.price >= $%d THEN 1 ELSE 0 END", argIndex))
			args = append(args, *extra.Best.MinPrice)
			argIndex++
		}

		if extra.Best.MaxPrice == nil {
			scoreParts = append(scoreParts, "1")
		} else {
			scoreParts = append(scoreParts, fmt.Sprintf("CASE WHEN Ad.price <= $%d THEN 1 ELSE 0 END", argIndex))
			args = append(args, *extra.Best.MaxPrice)
			argIndex++
		}

		if extra.Best.Radius == nil {
			scoreParts = append(scoreParts, "1")
		} else {
			scoreParts = append(scoreParts, fmt.Sprintf(`
			CASE
				WHEN Locality.latitude IS NULL OR Locality.longitude IS NULL THEN 0
				WHEN haversine_distance(Locality.latitude, Locality.longitude, $%d, $%d) <= $%d THEN 1
				ELSE 0
			END`, argIndex, argIndex+1, argIndex+2))
			args = append(args, extra.Latitude, extra.Longitude, *extra.Best.Radius)
			argIndex += 3
		}

		scoreExpr := "(" + strings.Join(scoreParts, " + ") + ")"
		query += fmt.Sprintf(" ORDER BY %s DESC, Ad.updated_at DESC ", scoreExpr)
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d;", argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

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
			lat      *float64
			lon      *float64
		)
		if err = rows.Scan(&row.ID, &row.OwnerID, &row.Status, &row.PhotoURL, &row.Title, &row.Description, &row.Price, &row.AnimalID, &row.BreedID, &row.Contacts, &row.CreatedAt, &row.UpdatedAt, &rowExtra.Username, &rowExtra.AnimalName, &rowExtra.BreedName, &rowExtra.LocalityName, &lat, &lon); err != nil {
			return result, errors.Wrap(err, "failed to parse ad")
		}
		result = append(result, ad.RespAd{Info: row, ExtraInfo: rowExtra})
	}

	return result, nil
}

func (repo *AdPostgres) SaveHistory(ctx context.Context, row ad.History) error {
	if _, err := repo.db.Exec(ctx, saveHistory, row.UserID, row.AnimalID, row.BreedID, row.MinPrice, row.MaxPrice, row.Radius, row.CreatedAt); err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return ad.ErrInvalidForeignKey
		}
		return errors.Wrap(err, "failed to save history")
	}

	return nil
}

func (repo *AdPostgres) GetHistory(ctx context.Context, userID uuid.UUID) (*ad.History, error) {
	result := ad.History{}
	if err := repo.db.QueryRow(ctx, getHistory, userID).Scan(&result.UserID, &result.AnimalID, &result.BreedID, &result.MinPrice, &result.MaxPrice, &result.Radius, &result.CreatedAt); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get history")
	}
	return &result, nil
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
