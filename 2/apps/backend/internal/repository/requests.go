package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

var retryStrategy = retry.Strategy{
	Attempts: 3,
	Delay:    time.Second,
	Backoff:  math.Exp(1),
}

func GetLongURL(db *dbpg.DB, shortURL string) (uuid.UUID, string, error) {
	query := `SELECT id, url FROM link WHERE short_url = $1`

	row, err := db.QueryWithRetry(context.Background(), retryStrategy, query, shortURL)
	if err != nil {
		return uuid.UUID{}, "", err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting longURL from DB")

	var url string
	var linkUUID uuid.UUID

	for row.Next() {
		zlog.Logger.Info().Msg("First row")

		err = row.Scan(&linkUUID, &url)
		if err != nil {
			return uuid.UUID{}, "", err
		}
	}

	zlog.Logger.Info().Msg(url)

	return linkUUID, url, nil
}

func InsertURL(db *dbpg.DB, url *URL) error {
	query := `INSERT INTO link VALUES ($1, $2, $3)`

	_, err := db.ExecWithRetry(context.Background(), retryStrategy, query, url.UUID, url.ShortURL, url.URL)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Inserting URL")

	return nil
}

func SaveInfo(db *dbpg.DB, url *URLInfo) error {
	query := `INSERT INTO link_following VALUES ($1, $2, $3, $4, $5)`

	_, err := db.ExecWithRetry(context.Background(), retryStrategy, query, url.UUID, url.LinkID, url.Time, url.UserAgent, url.IP)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Inserting url following information")

	return nil
}

func GetInfo(db *dbpg.DB, shortURL string) ([]URLInfo, error) {
	query := `SELECT l.id, l.link_id, l.time, l.user_agent, l.ip FROM link_following l
JOIN link ON link.id = l.link_id
WHERE link.short_url = $1`

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query, shortURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting url following information")

	allURLInfo := make([]URLInfo, 0)

	for rows.Next() {
		urlInfo := &URLInfo{}
		err = rows.Scan(
			&urlInfo.UUID,
			&urlInfo.LinkID,
			&urlInfo.Time,
			&urlInfo.UserAgent,
			&urlInfo.IP,
		)
		if err != nil {
			return nil, err
		}

		allURLInfo = append(allURLInfo, *urlInfo)
	}

	return allURLInfo, nil
}

func GetInfoWithGroup(db *dbpg.DB, shortURL, param string) ([]AnalyticsGroups, error) {
	query := fmt.Sprintf(`SELECT %s as parameter, COUNT(*) as visits, COUNT(DISTINCT ip) as unique_visitors FROM link_following l
	JOIN link ON link.id = l.link_id
	WHERE link.short_url = $1
	GROUP BY %s`, param, param)

	rows, err := db.QueryWithRetry(context.Background(), retryStrategy, query, shortURL)
	if err != nil {
		return nil, nil
	}

	defer func() {
		err = rows.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	zlog.Logger.Info().Msg("Getting url following information")

	allURLInfo := make([]AnalyticsGroups, 0)

	for rows.Next() {
		AnalyticsGroup := &AnalyticsGroups{}
		err = rows.Scan(
			&AnalyticsGroup.Parameter,
			&AnalyticsGroup.Visitors,
			&AnalyticsGroup.UniqueVisitors,
		)
		if err != nil {
			return nil, err
		}

		allURLInfo = append(allURLInfo, *AnalyticsGroup)
	}

	return allURLInfo, nil
}
