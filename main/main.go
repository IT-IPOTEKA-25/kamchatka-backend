package main

import (
	"context"
	"fmt"
	chatgpt2 "github.com/IT-IPOTEKA-25/kamchatka-backend/chatgpt"
	pb "github.com/IT-IPOTEKA-25/kamchatka-backend/generated/go"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

const (
	port = ":50051"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get database credentials from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbIP := os.Getenv("DB_IP")
	dbName := os.Getenv("DB_NAME")
	aiKey := os.Getenv("AI_KEY")

	// Connect to PostgreSQL database
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s/%s", dbUser, dbPass, dbIP, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// Ping the database to ensure the connection is established
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database!")

	// Create a new instance of the server, passing the database connection
	srv := NewServer(conn, chatgpt2.NewChatGpt(aiKey))

	// Set up a TCP listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()
	pb.RegisterKamchatkaServiceServer(s, srv)

	// Start the gRPC server
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
