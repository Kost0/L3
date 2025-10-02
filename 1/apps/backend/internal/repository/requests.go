package repository

import (
	"context"
	"log"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

var retryStrategy = retry.Strategy{
	Attempts: 3,
	Delay:    time.Second,
	Backoff:  1.5,
}

func GetNotifyByID(id string, db *dbpg.DB) (string, error) {
	query := `SELECT status FROM notify WHERE id = $1`

	ctx := context.Background()

	rows, err := db.QueryWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return "", err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			log.Printf("rows.Close(): %v", err)
		}
	}()

	var result string

	err = rows.Scan(&result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func CreateNotify(notify *Notify, db *dbpg.DB) error {
	query := `INSERT INTO notify (status, text, date) VALUES ($1, $2)`

	ctx := context.Background()

	status := "waiting"

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, status, notify.Text, notify.SendAt)
	if err != nil {
		return err
	}

	return nil
}

func DeleteNotifyByID(id string, db *dbpg.DB) error {
	query := `DELETE FROM notify WHERE id = $1`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return err
	}

	return nil
}
