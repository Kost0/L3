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

func GetAllOrders(db *dbpg.DB) ([]Order, error) {
	query := `
SELECT * FROM orders;
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

	orders := make([]Order, 0)

	for row.Next() {
		var order Order

		err = row.Scan(
			&order.OrderID,
			&order.Title,
			&order.Cost,
			&order.Items,
			&order.Category,
			&order.Date,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func CreateOrder(db *dbpg.DB, order *Order) error {
	query := `
INSERT INTO orders VALUES ($1, $2, $3, $4, $5, $6);
`

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = tx.ExecContext(ctx, query, order.OrderID, order.Title, order.Cost, order.Items, order.Category, order.Date)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func UpdateOrder(db *dbpg.DB, order *Order) error {
	query := `
UPDATE orders SET (title = $1, cost = $2, items = $3, category = $4, date = $5) WHERE uuid = $6;
`

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = tx.ExecContext(ctx, query, order.Title, order.Cost, order.Items, order.Category, order.Date, order.OrderID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func DeleteOrder(db *dbpg.DB, id string) error {
	query := `
DELETE FROM orders WHERE uuid = $1;
`

	tx, err := db.Master.Begin()
	if err != nil {
		return err
	}

	ctx := context.Background()

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

func GetSum(db *dbpg.DB, column string, from, to string) (int, error) {
	query := fmt.Sprintf(`
SELECT SUM(%s) FROM orders
WHERE date BETWEEN $1 AND $2;
`, column)

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var sum int

	for row.Next() {
		err = row.Scan(&sum)
		if err != nil {
			return 0, err
		}
	}

	return sum, nil
}

func GetAvg(db *dbpg.DB, column string, from, to string) (float64, error) {
	query := fmt.Sprintf(`
SELECT AVG(%s) FROM orders
WHERE date BETWEEN $1 AND $2;
`, column)

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var avg float64

	for row.Next() {
		err = row.Scan(&avg)
		if err != nil {
			return 0, err
		}
	}

	return avg, nil
}

func GetCount(db *dbpg.DB, from, to string) (int, error) {
	query := `
SELECT COUNT(*) FROM orders
WHERE date BETWEEN $1 AND $2;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var count int

	for row.Next() {
		err = row.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func GetMedian(db *dbpg.DB, column string, from, to string) (float64, error) {
	query := fmt.Sprintf(`
SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY %s) FROM orders
WHERE date BETWEEN $1 AND $2;
`, column)

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var median float64

	for row.Next() {
		err = row.Scan(&median)
		if err != nil {
			return 0, err
		}
	}

	return median, nil
}

func GetPercentile(db *dbpg.DB, column string, from, to string) (float64, error) {
	query := fmt.Sprintf(`
SELECT PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY %s) FROM orders
WHERE date BETWEEN $1 AND $2;
`, column)

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var per float64

	for row.Next() {
		err = row.Scan(&per)
		if err != nil {
			return 0, err
		}
	}

	return per, nil
}

func GetAllCategories(db *dbpg.DB) ([]string, error) {
	query := `
SELECT DISTINCT category FROM orders;
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

	categories := make([]string, 0)

	for row.Next() {
		category := ""

		err = row.Scan(&category)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func GetOrdersByCategory(db *dbpg.DB, category string) ([]Order, error) {
	query := `
SELECT * FROM orders
WHERE category = $1;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, category)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	orders := make([]Order, 0)

	for row.Next() {
		order := Order{}

		err = row.Scan(
			&order.OrderID,
			&order.Title,
			&order.Cost,
			&order.Items,
			&order.Category,
			&order.Date,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}
