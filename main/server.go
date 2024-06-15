package main

import (
	"context"
	"encoding/json"
	pb "github.com/IT-IPOTEKA-25/kamchatka-backend/generated/go"
	"github.com/jackc/pgx/v4"
)

type Server struct {
	pb.UnimplementedKamchatkaServiceServer
	conn *pgx.Conn
}

func NewServer(conn *pgx.Conn) *Server {
	return &Server{conn: conn}
}

type Coordinates struct {
	Name string
	Dot  []string
}

func (s *Server) GetTerritoryCoordinates(ctx context.Context, req *pb.GetRouteCoordinatesRequest) (*pb.GetRouteCoordinatesResponse, error) {
	var dbResult string
	var coordinates []Coordinates
	err := s.conn.QueryRow(ctx, "select coordinates from kamchatka.security_territories_coordinates where territory_id = $1", req.Id).Scan(&dbResult)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(dbResult), &coordinates)
	if err != nil {
		return nil, err
	}
	var pbCoordinates []*pb.Coordinate
	for _, coordinate := range coordinates {
		pbCoordinates = append(pbCoordinates, &pb.Coordinate{
			Name: coordinate.Name,
			Dot:  coordinate.Dot,
		})
	}
	return &pb.GetRouteCoordinatesResponse{
		Coordinates: pbCoordinates,
	}, nil
}

type RecreationalCapacity struct {
	LengthDtp              float32
	DaytimeTs              float32
	DistanceDgp            float32
	AverageTimeTdp         float32
	AverageHumansInGroupGs int32
	AverageDaysOnPathTp    float32
	ResultCapacity         float32
}

func (s *Server) GetRecreationalCapacity(ctx context.Context, req *pb.GetRecreationalCapacityRequest) (*pb.GetRecreationalCapacityResponse, error) {
	var resultCapacity RecreationalCapacity
	err := s.conn.QueryRow(ctx, "select length_dtp, daytime_ts, distance_dgp, average_time_tdp, average_humans_in_group_gs, average_days_on_path_tp, result_capacity from kamchatka.recreational_capacity where territory_id = &1", req.Id).Scan(&resultCapacity)
	if err != nil {
		return nil, err
	}
	return &pb.GetRecreationalCapacityResponse{
		Length:             resultCapacity.LengthDtp,
		Daytime:            resultCapacity.DaytimeTs,
		Distance:           resultCapacity.DistanceDgp,
		AverageTime:        resultCapacity.AverageTimeTdp,
		AverageHumans:      resultCapacity.AverageHumansInGroupGs,
		AverageDays:        resultCapacity.AverageDaysOnPathTp,
		RecreationalResult: resultCapacity.ResultCapacity,
	}, nil
}

func (s *Server) AddAlert(ctx context.Context, req *pb.AddAlertRequest) (*pb.StringResultResponse, error) {
	_, err := s.conn.Exec(ctx, "insert into kamchatka.users_alerts(user_id, description, url) VALUES ($1, $2, $3)", req.UserId, req.Description, req.ImageUrl)
	if err != nil {
		return nil, err
	}
	return &pb.StringResultResponse{
		Result: "Successfully added alert",
	}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var id int64
	err := s.conn.QueryRow(ctx, "insert into kamchatka.users(name, phone) VALUES ($1, $2)", req.Name, req.Phone).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{
		Id: id,
	}, nil
}

func (s *Server) UpdateRecreationalCapacity(ctx context.Context, req *pb.UpdateRecreationalCapacityRequest) (*pb.UpdateRecreationalCapacityResponse, error) {
	result := (req.Length * req.Daytime) / (req.AverageDays * req.AverageTime * float32(req.AverageHumans) * req.Distance)
	_, err := s.conn.Exec(ctx, "update kamchatka.recreational_capacity set length_dtp = $1, daytime_ts = $2, distance_dgp = $3, average_time_tdp = $4, average_humans_in_group_gs = $5, average_days_on_path_tp = $6, result_capacity = $7 where territory_id = $8",
		req.Length, req.Daytime, req.Distance, req.AverageTime, req.AverageHumans, req.AverageDays, result, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateRecreationalCapacityResponse{
		RecreationalResult: result,
	}, nil
}

type Groups struct {
	ID   int64
	Name string
}

func (s *Server) GetGroups(ctx context.Context, _ *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	var groups []Groups
	err := s.conn.QueryRow(ctx, "select * from kamchatka.security_territories_groups").Scan(&groups)
	if err != nil {
		return nil, err
	}
	var pbGroups []*pb.Group
	for _, group := range groups {
		pbGroups = append(pbGroups, &pb.Group{
			Id:   group.ID,
			Name: group.Name,
		})
	}
	return &pb.GetGroupsResponse{
		Groups: pbGroups,
	}, nil
}

func (s *Server) GetTerritory(ctx context.Context, req *pb.GetTerritoryRequest) (*pb.GetTerritoryResponse, error) {
	var id int64
	var name string
	var currentWorkload int64
	var routeOpen bool
	err := s.conn.QueryRow(ctx, "select id, name, current_workload, route_open from kamchatka.security_territories where group_id = $1 and has_data = true", req.Id).
		Scan(&id, &name, &currentWorkload, &routeOpen)
	if err != nil {
		return nil, err
	}
	return &pb.GetTerritoryResponse{
		Id:              id,
		Name:            name,
		CurrentWorkload: currentWorkload,
		RouteOpen:       routeOpen,
	}, nil
}
