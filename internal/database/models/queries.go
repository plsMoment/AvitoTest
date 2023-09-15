package models

import (
	"AvitoTest/internal/config"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

type Storage struct {
	db *pgxpool.Pool
}

type Segment struct {
	Id   string
	Slug string
}

func New(cfg *config.DB) (*Storage, error) {
	scope := "database.models.New"
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.Username, os.Getenv("DB_PASSWORD"), cfg.Host, cfg.Port, cfg.Name,
	)

	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", scope, err)
	}

	_, err = db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS "segments" (
		"id" uuid PRIMARY KEY,
		"slug" varchar NOT NULL UNIQUE);
	CREATE INDEX IF NOT EXISTS "idx_name" ON "segments"("slug");
	CREATE TABLE IF NOT EXISTS "user_segments" (
		"user_id" uuid NOT NULL,
		"segment_id" uuid NOT NULL,
		FOREIGN KEY ("segment_id") REFERENCES "segments" ("id") ON DELETE CASCADE,
		PRIMARY KEY ("user_id", "segment_id"));
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", scope, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) CreateSegment(slug string) error {
	scope := "database.models.CreateSegment"

	id := uuid.New()
	_, err := s.db.Exec(context.Background(), "INSERT INTO segments (id, slug) VALUES ($1, $2)", id, slug)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}

	return nil
}

func (s *Storage) DeleteSegment(slug string) error {
	scope := "database.models.DeleteSegment"

	res, err := s.db.Exec(context.Background(), "DELETE FROM segments WHERE slug = $1", slug)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("such segment doesn't exist")
	}

	return nil
}

func (s *Storage) GetUserSegments(userId uuid.UUID) ([]string, error) {
	scope := "database.models.GetUserSegments"

	rows, err := s.db.Query(
		context.Background(),
		`SELECT slug FROM user_segments us INNER JOIN segments s ON us.segment_id = s.id WHERE us.user_id = $1`,
		userId,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", scope, err)
	}
	defer rows.Close()

	var res []string
	for rows.Next() {
		var slug string
		if err = rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("%s: %w", scope, err)
		}
		res = append(res, slug)
	}

	return res, nil
}

func (s *Storage) ChangeUserSegments(userId uuid.UUID, addSlugs []string, deleteSlugs []string) error {
	scope := "database.models.ChangeUserSegments"
	ctx := context.Background()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}
	defer tx.Rollback(ctx)

	if len(addSlugs) != 0 {
		err = AddUserSegments(tx, userId, addSlugs)
		if err != nil {
			return err
		}
	}

	if len(deleteSlugs) != 0 {
		err = DeleteUserSegments(tx, userId, deleteSlugs)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}

	return nil
}

func AddUserSegments(tx pgx.Tx, userId uuid.UUID, addSlugs []string) error {
	scope := "database.models.AddUserSegments"

	rows, err := tx.Query(context.Background(), "SELECT * FROM segments WHERE slug = ANY ($1)", addSlugs)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}
	defer rows.Close()
	segments, err := pgx.CollectRows(rows, pgx.RowToStructByName[Segment])
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}
	if len(addSlugs) != len(segments) {
		return fmt.Errorf(
			"%s: some segments was not found, number: %d",
			scope, len(addSlugs)-len(segments),
		)
	}

	batch := &pgx.Batch{}
	query := "INSERT INTO user_segments (user_id, segment_id) VALUES ($1, $2)"
	for _, segment := range segments {
		batch.Queue(query, userId, segment.Id)
	}
	res := tx.SendBatch(context.Background(), batch)
	defer res.Close()

	var errs error
	for _, segment := range segments {
		_, err := res.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				errs = errors.Join(errs, fmt.Errorf("user already has segment %s", segment.Slug))
			}
			return errors.Join(errs, fmt.Errorf("%s: %w", scope, err))
		}
	}

	return errs
}

func DeleteUserSegments(tx pgx.Tx, userId uuid.UUID, deleteSlugs []string) error {
	scope := "database.models.DeleteUserSegments"

	rows, err := tx.Query(context.Background(), "SELECT * FROM segments WHERE slug = ANY ($1)", deleteSlugs)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}
	defer rows.Close()

	segments, err := pgx.CollectRows(rows, pgx.RowToStructByName[Segment])
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}
	if len(deleteSlugs) != len(segments) {
		return fmt.Errorf(
			"%s: some segments was not found, number: %d",
			scope, len(deleteSlugs)-len(segments),
		)
	}

	var ids []string
	for _, segment := range segments {
		ids = append(ids, segment.Id)
	}

	_, err = tx.Exec(context.Background(),
		"DELETE FROM user_segments WHERE user_id = $1 AND segment_id = ANY ($2)", userId, ids,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", scope, err)
	}

	return nil
}
