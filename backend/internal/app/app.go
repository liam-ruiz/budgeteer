package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/liam-ruiz/budget/internal/api"
	"github.com/liam-ruiz/budget/internal/bank_accounts"
	"github.com/liam-ruiz/budget/internal/bank_accounts/plaid_items"
	"github.com/liam-ruiz/budget/internal/budgets"
	"github.com/liam-ruiz/budget/internal/config"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/dependencies"
	"github.com/liam-ruiz/budget/internal/plaid"
	"github.com/liam-ruiz/budget/internal/transactions"
	"github.com/liam-ruiz/budget/internal/users"
	
)

func Run(cfg *config.Config) error {
	db, err := initDB(cfg.DBUrl)
	if err != nil {
		return err
	}
	defer db.Close()

	queries := sqlcdb.New(db)

	// reset database if requested
	if cfg.ResetDB {
		if err := resetDatabase(db); err != nil {
			return fmt.Errorf("failed to reset database: %w", err)
		}
	}

	// Build the container
	plaidItemService := plaid_items.NewService(queries)
	plaidClient := plaid.NewPlaidClient(cfg.PlaidClientID, cfg.PlaidSecret, cfg.PlaidEnv)
	plaidService := plaid.NewService(plaidClient)
	cont := dependencies.NewContainer(
		users.NewService(users.NewRepository(queries)),
		bank_accounts.NewService(bank_accounts.NewRepository(queries), plaidItemService, plaidService),
		budgets.NewService(budgets.NewRepository(queries)),
		transactions.NewService(transactions.NewRepository(queries)),
		plaidService,
		cfg,
		plaidItemService,
	)



	// Now NewHandler only takes the container
	handler := api.NewHandler(cont)

	log.Println("Application Started.")
	return http.ListenAndServe(":"+cfg.Port, handler.Routes())
}

func initDB(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbUrl)
	return db, err
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	// Path to your migration files (relative to the binary)
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	// Apply all "up" migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func resetDatabase(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	// 1. Rollback all migrations
	// This will execute every .down.sql file
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to roll back: %w", err)
	}

	// 2. Re-apply all migrations
	// This will execute every .up.sql file
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to re-apply: %w", err)
	}

	return nil
}
