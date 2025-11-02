package repository

import (
	"context"
	"fmt"
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

func CreateItem(db *dbpg.DB, item *Item, roleID string) error {
	query := `
INSERT INTO items VALUES ($1, $2, $3, $4, $5, $6);
`

	queryForRole := fmt.Sprintf(`SET LOCAL app.current_role_id = %s`, roleID)

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = tx.ExecContext(ctx, queryForRole)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, query, item.UUID, item.Title, item.Price, item.Category, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetAllItems(db *dbpg.DB) ([]Item, error) {
	query := `
SELECT * FROM items;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	items := make([]Item, 0)

	for row.Next() {
		var item Item

		err = row.Scan(
			&item.UUID,
			&item.Title,
			&item.Price,
			&item.Category,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func UpdateItem(db *dbpg.DB, item *ItemDTO, id string, roleID string) error {
	query := `
UPDATE items SET title = $1, price = $2, category = $3, updated_at = $4 WHERE uuid=$5;
`

	queryForRole := fmt.Sprintf(`SET LOCAL app.current_role_id = %s`, roleID)

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = tx.ExecContext(ctx, queryForRole)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, query, item.Title, item.Price, item.Category, time.Now(), id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func DeleteItem(db *dbpg.DB, id string, roleID string) error {
	query := `
DELETE FROM items WHERE uuid = $1;
`

	queryForRole := fmt.Sprintf(`SET LOCAL app.current_role_id = %s`, roleID)

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = tx.ExecContext(ctx, queryForRole)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetHistory(db *dbpg.DB, item string) ([]History, error) {
	query := `
SELECT item_id, operation_type, r.name AS role_name, changed_at
FROM history
JOIN roles AS r ON r.id = history.editor_id
WHERE history.item_id = $1;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, item)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	allHistory := make([]History, 0)

	for row.Next() {
		var history History

		err = row.Scan(
			&history.ItemID,
			&history.OperationType,
			&history.RoleName,
			&history.ChangedAt,
			//&history.Difference,
		)
		if err != nil {
			return nil, err
		}

		allHistory = append(allHistory, history)
	}

	return allHistory, nil
}
