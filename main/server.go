package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IT-IPOTEKA-25/kamchatka-backend/chatgpt"
	pb "github.com/IT-IPOTEKA-25/kamchatka-backend/generated/go"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"strings"
	"unicode"
)

type Server struct {
	pb.UnimplementedKamchatkaServiceServer
	conn    *pgx.Conn
	chatgpt *chatgpt.ChatGpt
}

func NewServer(conn *pgx.Conn, chatgpt *chatgpt.ChatGpt) *Server {
	return &Server{
		conn:    conn,
		chatgpt: chatgpt,
	}
}

type Coordinates struct {
	Name string
	Dot  []string
}

func parseDMSString(dmsString string) (int, int, int, error) {
	dmsString = strings.TrimLeftFunc(dmsString, func(r rune) bool {
		return unicode.IsLetter(r)
	})
	var degrees, minutes, seconds int
	_, err := fmt.Sscanf(dmsString, "%d°%d'%d\"", &degrees, &minutes, &seconds)
	if err != nil {
		return 0, 0, 0, err
	}
	return degrees, minutes, seconds, nil
}

func convertDMSToDD(data string) float32 {
	degrees, minutes, seconds, _ := parseDMSString(data)
	decimalDegrees := float64(degrees) + float64(minutes)/60.0 + float64(seconds)/3600.0
	return float32(math.Round(decimalDegrees*10000) / 10000) // округление до 4 знаков после запятой
}

func (s *Server) GetRouteCoordinates(ctx context.Context, req *pb.GetRouteCoordinatesRequest) (*pb.GetRouteCoordinatesResponse, error) {
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
			Dot: []float32{
				convertDMSToDD(coordinate.Dot[0]),
				convertDMSToDD(coordinate.Dot[1]),
			},
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
	err := s.conn.QueryRow(ctx, "select length_dtp, daytime_ts, distance_dgp, average_time_tdp, average_humans_in_group_gs, average_days_on_path_tp, result_capacity from kamchatka.recreational_capacity where territory_id = $1", req.Id).Scan(
		&resultCapacity.LengthDtp,
		&resultCapacity.DaytimeTs,
		&resultCapacity.DistanceDgp,
		&resultCapacity.AverageTimeTdp,
		&resultCapacity.AverageHumansInGroupGs,
		&resultCapacity.AverageDaysOnPathTp,
		&resultCapacity.ResultCapacity,
	)
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
	//TODO: Temporary disabled
	//isTrash, err := s.chatgpt.Prompt(req.ImageUrl)
	//if err != nil {
	//	return nil, err
	//}
	//if !isTrash{
	//	return nil, errors.New("not found any trash on image")
	//}
	_, sqlErr := s.conn.Exec(ctx, "insert into kamchatka.users_alerts(user_id, description, url) VALUES ($1, $2, $3)", req.UserId, req.Description, req.ImageUrl)
	if sqlErr != nil {
		return nil, sqlErr
	}
	return &pb.StringResultResponse{
		Result: "Successfully added alert",
	}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var id int64
	err := s.conn.QueryRow(ctx, "INSERT INTO kamchatka.users(name, phone) VALUES ($1, $2) RETURNING id", req.Name, req.Phone).Scan(&id)
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
	rows, err := s.conn.Query(ctx, "select id, name from kamchatka.security_territories_groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Groups
	for rows.Next() {
		var group Groups
		err = rows.Scan(&group.ID, &group.Name)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
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

type Territory struct {
	Id              int64
	Name            string
	CurrentWorkload int64
	RouteOpen       bool
}

func (s *Server) GetGroupTerritories(ctx context.Context, req *pb.GetGroupTerritoriesRequest) (*pb.GetGroupTerritoriesResponse, error) {
	rows, err := s.conn.Query(ctx, "select id, name, current_workload, route_open from kamchatka.security_territories where group_id = $1 and has_data = true", req.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var territories []Territory
	for rows.Next() {
		var territory Territory
		err = rows.Scan(&territory.Id, &territory.Name, &territory.CurrentWorkload, &territory.RouteOpen)
		if err != nil {
			return nil, err
		}
		territories = append(territories, territory)
	}
	var pbTerritories []*pb.Territory
	for _, territory := range territories {
		pbTerritories = append(pbTerritories, &pb.Territory{
			Id:              territory.Id,
			Name:            territory.Name,
			CurrentWorkload: territory.CurrentWorkload,
			RouteOpen:       territory.RouteOpen,
		})
	}
	return &pb.GetGroupTerritoriesResponse{
		Territories: pbTerritories,
	}, nil
}

type SatelliteAlerts struct {
	Image       string
	Category    string
	Time        timestamppb.Timestamp
	Coordinates string
}

func (s *Server) GetSatelliteAlerts(ctx context.Context, _ *pb.GetSatelliteAlertsRequest) (*pb.GetSatelliteAlertsResponse, error) {
	rows, err := s.conn.Query(ctx, "select image, category, time, coordinates from kamchatka.satellite_alerts where handled is not true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []*SatelliteAlerts
	for rows.Next() {
		var alert SatelliteAlerts
		err = rows.Scan(&alert.Image, &alert.Category, &alert.Time, &alert.Coordinates)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, &alert)
	}
	var satelliteAlerts []*pb.SatelliteAlert
	for _, alert := range alerts {
		satelliteAlerts = append(satelliteAlerts, &pb.SatelliteAlert{
			Image:       alert.Image,
			Category:    alert.Category,
			Time:        alert.Time.String(),
			Coordinates: alert.Coordinates,
		})
	}
	return &pb.GetSatelliteAlertsResponse{
		SatelliteAlerts: satelliteAlerts,
	}, nil
}
