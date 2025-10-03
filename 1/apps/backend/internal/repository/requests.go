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

var Deleted = make(map[string]struct{})

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
	query := `INSERT INTO notify (id, status, text, send_at, email, tg_user) 
VALUES ($1, $2, $3, $4, $5, $6)`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, notify.ID, notify.Status, notify.Text, notify.SendAt, notify.Email, notify.TGUser)
	if err != nil {
		return err
	}

	return nil
}

func DeleteNotifyByID(id string, db *dbpg.DB) error {
	query := `DELETE FROM notify WHERE id = $1`

	Deleted[id] = struct{}{}

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return err
	}

	return nil
}

func CheckMigrations(db *dbpg.DB) error {
	query := `SELECT * FROM notify`

	ctx := context.Background()

	_, err := db.QueryWithRetry(ctx, retryStrategy, query)
	if err != nil {
		return err
	}

	return nil
}
