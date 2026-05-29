package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Zheng5005/onemorerep/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	fmt.Println("seed completed successfully")
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://onemorerep:onemorerep@localhost:5432/onemorerep?sslmode=disable"
	}

	ctx := context.Background()
	db, err := store.NewDB(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	fmt.Println("connected to database")
	fmt.Println("sample exercises can be inserted here via sqlc queries or raw SQL")
	fmt.Println("seed completed")
	return nil
}
