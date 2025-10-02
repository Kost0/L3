package repository

import (
	"fmt"
	"os"
	"time"

	"github.com/wb-go/wbf/dbpg"
)

func ConnectDB() (*dbpg.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Second,
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := dbpg.New(connStr, []string{}, opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}
