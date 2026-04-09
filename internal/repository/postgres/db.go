package postgres

import (
	"log/slog"

	"github.com/dev-32/auth-service/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib" // ✅ pgx driver — handles UUID natively
	"github.com/jmoiron/sqlx"
)

func Connect(cfg *config.Config) *sqlx.DB {
	db, err := sqlx.Connect("pgx", cfg.DBDSN())
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err.Error())
		panic(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	slog.Info("connected to postgres")
	return db
}
