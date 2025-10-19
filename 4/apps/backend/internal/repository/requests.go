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

func InsertPhotoData(db *dbpg.DB, photo *Photo) error {
	query := `INSERT INTO photos(uuid, status) VALUES ($1, $2)`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, photo.UUID, photo.Status)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Photo inserted in db")

	return nil
}

func UpdatePhotoData(db *dbpg.DB, photo *Photo) error {
	query := `UPDATE photos SET status = $1 WHERE uuid = $2`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, photo.Status, photo.UUID)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Photo updated in db")

	return nil
}

func GetStatus(db *dbpg.DB, id string) (string, error) {
	query := `SELECT status FROM photos WHERE uuid = $1`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return "", err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	status := ""

	for row.Next() {
		zlog.Logger.Info().Msg("First row")

		err = row.Scan(&status)
		if err != nil {
			return "", err
		}
	}

	zlog.Logger.Info().Msg(status)

	return status, nil
}
