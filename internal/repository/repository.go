package repository

import (
	"context"
	"errors"
	"time"
	"github.com/jackc/pgx"
	"log"
	"github.com/kujoki/loyalty-service/internal/model"
)

type PostgresRepository struct {
    Pool *pgx.ConnPool
}

func parseConDSN(connectionDSN string) (*pgx.ConnPoolConfig) {
	config, err := pgx.ParseConnectionString(connectionDSN)
	if err != nil {
		log.Printf("failed to parse connection string: %v", err)
		return nil
	}
	log.Println("successful parse connection string")
	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 10,
	}
	return &poolConfig
}

func NewPostgresRepository(connectionDSN string) (*PostgresRepository, error) {
	config := parseConDSN(connectionDSN)
	pool, err := pgx.NewConnPool(*config)
	if err != nil {
		log.Printf("failed to create pool: %v", err)
		return nil, err
	}
	log.Println("create connection pool")
	return &PostgresRepository{Pool: pool}, nil
}

func (r *PostgresRepository) SetAuth(ctx context.Context, login string, pwd string) error {
	var loginDB string
    err := r.Pool.QueryRow("SELECT login FROM users WHERE login=$1", login).Scan(&loginDB)
    if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("this login already exists: %v \n", err)
        return model.ErrLoginExists
    }
	log.Println("try to set new auth")
	_, er := r.Pool.Exec(`
        INSERT INTO users (login, password)
        VALUES ($1, $2)
    `, login, pwd)
    return er
}

func (r *PostgresRepository) GetHashPassword(ctx context.Context, login string) (string, error) {
	var hashedPassword string
	err := r.Pool.QueryRow("SELECT password FROM users WHERE login=$1", login).Scan(&hashedPassword)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}