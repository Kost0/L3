package repository

import (
	"context"
	"math"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

var retryStrategy = retry.Strategy{
	Attempts: 3,
	Delay:    time.Second,
	Backoff:  math.Exp(1),
}

func InsertComment(db *dbpg.DB, comment *Comment) error {
	query := `INSERT INTO comment(id, text, parent) VALUES ($1, $2, $3)`

	_, err := db.ExecWithRetry(context.Background(), retryStrategy, query, comment.UUID, comment.Text, comment.Parent)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Inserted comment")

	return nil
}

func SelectRootComments(db *dbpg.DB) ([]*Comment, error) {
	query := `SELECT * FROM comment WHERE parent IS NULL`

	zlog.Logger.Info().Msg("Getting comments...")

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	res := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}
		err = rows.Scan(
			&comment.UUID,
			&comment.Text,
			&comment.Parent,
			&comment.Vector,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, comment)
	}

	zlog.Logger.Info().Msgf("Got %d comments", len(res))

	return res, nil
}

func SelectComments(db *dbpg.DB, id string) ([]*Comment, error) {
	if id == "null" {
		return []*Comment{}, nil
	}

	query := `SELECT * FROM comment
	WHERE comment.parent = $1`

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query, id)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting comments")

	res := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}
		err = rows.Scan(
			&comment.UUID,
			&comment.Text,
			&comment.Parent,
			&comment.Vector,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, comment)
	}

	return res, nil
}

func DeleteComments(db *dbpg.DB, id string) error {
	query := `DELETE FROM comment WHERE comment.id = $1`

	_, err := db.ExecWithRetry(context.Background(), retryStrategy, query, id)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Comments deleted")

	return nil
}

func SearchComments(db *dbpg.DB, keyword string) ([]*Comment, error) {
	query := `SELECT * FROM comment
WHERE search_vector @@ to_tsquery('russian', $1)
ORDER BY ts_rank(search_vector, to_tsquery('russian', $1)) DESC`

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query, keyword)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting comments")

	res := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}
		err = rows.Scan(
			&comment.UUID,
			&comment.Text,
			&comment.Parent,
			&comment.Vector,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, comment)
	}

	return res, nil
}

func SelectPage(db *dbpg.DB, page int, sort string) ([]*Comment, error) {
	query := `SELECT * FROM comment
WHERE parent IS NULL
`

	if sort == "desc" {
		query += ` ORDER BY text DESC`
	} else if sort == "asc" {
		query += ` ORDER BY text`
	}

	query += ` LIMIT 10 OFFSET ($1 - 1) * 10`

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query, page)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting comments")

	res := make([]*Comment, 0)

	for rows.Next() {
		comment := &Comment{}
		err = rows.Scan(
			&comment.UUID,
			&comment.Text,
			&comment.Parent,
			&comment.Vector,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, comment)
	}

	return res, nil
}

func CountPages(db *dbpg.DB) (int, error) {
	query := `SELECT COUNT(*) FROM comment`

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query)
	if err != nil {
		return 0, err
	}

	var count int

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	return count/10 + 1, nil
}
