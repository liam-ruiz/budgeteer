package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/liam-ruiz/budget/internal/api"
	"github.com/liam-ruiz/budget/internal/bank_accounts"
	"github.com/liam-ruiz/budget/internal/plaid_items"
	"github.com/liam-ruiz/budget/internal/budgets"
	"github.com/liam-ruiz/budget/internal/config"
	"github.com/liam-ruiz/budget/internal/db/sqlcdb"
	"github.com/liam-ruiz/budget/internal/dependencies"
	"github.com/liam-ruiz/budget/internal/plaid"
	"github.com/liam-ruiz/budget/internal/transactions"
	"github.com/liam-ruiz/budget/internal/users"
)

func Run(cfg *config.Config) error {
	dbPool, err := pgxpool.New(context.Background(), cfg.DBUrl)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer dbPool.Close()

    queries := sqlcdb.New(dbPool)

	// reset database if requested
	if cfg.ResetDB {
		if err := resetDatabase(dbPool); err != nil {
			return fmt.Errorf("failed to reset database: %w", err)
		}
	}

	// Build the container
	plaidItemService := plaid_items.NewService(queries)
	plaidClient := plaid.NewPlaidClient(cfg.PlaidClientID, cfg.PlaidSecret, cfg.PlaidEnv)
	plaidService := plaid.NewService(plaidClient)
	transactionsService := transactions.NewService(transactions.NewRepository(queries))
	cont := dependencies.NewContainer(
		users.NewService(users.NewRepository(queries)),
		bank_accounts.NewService(bank_accounts.NewRepository(queries), plaidItemService, plaidService, transactionsService),
		budgets.NewService(budgets.NewRepository(queries)),
		transactionsService,
		plaidService,
		cfg,
		plaidItemService,
	)



	// Now NewHandler only takes the container
	handler := api.NewHandler(cont, cfg.BaseURL, cfg.WebhookURL)

	log.Println("Application Started.")
	return http.ListenAndServe(":"+cfg.Port, handler.Routes())
}

func runMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	// Path to your migration files (relative to the binary)
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"pgx", driver)
	if err != nil {
		return err
	}

	// Apply all "up" migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func resetDatabase(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"pgx", driver)
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
