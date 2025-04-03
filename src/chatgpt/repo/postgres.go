package repo

import (
	"context"
	goerrors "errors"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/chatgpt"
)

const (
	getDescriptions   = `SELECT id, color, created_at, updated_at FROM GptDescription;`
	getDescription    = `SELECT id, color, created_at, updated_at FROM GptDescription WHERE id = $1;`
	createDescription = `INSERT INTO GptDescription(id, color, created_at, updated_at) VALUES ($1, $2, $3, $4);`
	updateDescription = `UPDATE GptDescription SET color = $1, updated_at = $2 WHERE id = $3;`
	deleteDescription = `DELETE FROM GptDescription WHERE id = $1;`
)

type DescriptionPostgres struct {
	db pgxtype.Querier
}

func NewDescriptionPostgres(db pgxtype.Querier) *DescriptionPostgres {
	return &DescriptionPostgres{db: db}
}

func (repo *DescriptionPostgres) GetDescriptions(ctx context.Context) ([]chatgpt.PostgresDescription, error) {
	result := make([]chatgpt.PostgresDescription, 0)

	query, err := repo.db.Query(ctx, getDescriptions)
	if err != nil {
		return result, errors.Wrap(err, "failed to get descriptions from postgres")
	}
	defer query.Close()

	for query.Next() {
		var row chatgpt.PostgresDescription
		if err = query.Scan(&row.ID, &row.Color, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return result, errors.Wrap(err, "failed to parse description")
		}
		result = append(result, row)
	}

	return result, nil
}

func (repo *DescriptionPostgres) GetDescription(ctx context.Context, id uuid.UUID) (chatgpt.PostgresDescription, error) {
	result := chatgpt.PostgresDescription{}
	if err := repo.db.QueryRow(ctx, getDescription, id).Scan(&result.ID, &result.Color, &result.CreatedAt, &result.UpdatedAt); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return chatgpt.PostgresDescription{}, chatgpt.ErrDescriptionNotFound
		}
		return chatgpt.PostgresDescription{}, errors.Wrap(err, "failed to get description from postgres")
	}
	return result, nil
}

func (repo *DescriptionPostgres) CreateDescription(ctx context.Context, description chatgpt.PostgresDescription) error {
	if _, err := repo.db.Exec(ctx, createDescription, description.ID, description.Color, description.CreatedAt, description.UpdatedAt); err != nil {
		return errors.Wrap(err, "failed to create description in postgres")
	}
	return nil
}

func (repo *DescriptionPostgres) UpdateDescription(ctx context.Context, description chatgpt.PostgresDescription) error {
	if _, err := repo.db.Exec(ctx, updateDescription, description.Color, description.UpdatedAt, description.ID); err != nil {
		return errors.Wrap(err, "failed to update description in postgres")
	}
	return nil
}

func (repo *DescriptionPostgres) DeleteDescription(ctx context.Context, id uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, deleteDescription, id); err != nil {
		return errors.Wrap(err, "failed to delete description from postgres")
	}
	return nil
}
