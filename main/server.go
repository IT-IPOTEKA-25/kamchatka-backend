package main

import (
	"context"
	pb "github.com/IT-IPOTEKA-25/kamchatka-backend/generated/go"
	"github.com/jackc/pgx/v4"
)

type Server struct {
	pb.UnimplementedKamchatkaServiceServer
	DBConn *pgx.Conn
}

func NewServer(conn *pgx.Conn) *Server {
	return &Server{DBConn: conn}
}

func (s *Server) GetTerritoryCoordinates(ctx context.Context, req *pb.GetTerritoryCoordinatesRequest) (*pb.GetTerritoryCoordinatesResponse, error) {
	// Implement your logic here
	return &pb.GetTerritoryCoordinatesResponse{}, nil
}

func (s *Server) GetRecreationalCapacity(ctx context.Context, req *pb.GetRecreationalCapacityRequest) (*pb.GetRecreationalCapacityResponse, error) {
	// Implement your logic here
	return &pb.GetRecreationalCapacityResponse{}, nil
}

func (s *Server) AddAlert(ctx context.Context, req *pb.AddAlertRequest) (*pb.StringResultResponse, error) {
	// Implement your logic here
	return &pb.StringResultResponse{}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Implement your logic here
	return &pb.CreateUserResponse{}, nil
}

func (s *Server) UpdateRecreationalCapacity(ctx context.Context, req *pb.UpdateRecreationalCapacityRequest) (*pb.UpdateRecreationalCapacityResponse, error) {
	// Implement your logic here
	return &pb.UpdateRecreationalCapacityResponse{}, nil
}
