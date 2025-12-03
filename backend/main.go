package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Database connection string
	databaseURL := "postgres://srini:srini@4614@localhost:5432/realtime_chat?sslmode=disable"

	// Create connection pool
	dbpool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// Test the connection
	err = dbpool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to the database!")

	// TODO: Add your application logic here
}
