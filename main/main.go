package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbIP := os.Getenv("DB_IP")
	dbName := os.Getenv("DB_NAME")

	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s/%s", dbUser, dbPass, dbIP, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")
}
