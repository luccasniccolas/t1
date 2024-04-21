package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pb "github.com/luccasniccolas/t1/proto"
	"google.golang.org/grpc"
)

var dbpool *pgxpool.Pool

func NewexampleServer() *server {
	return &server{}
}

type server struct {
	conn *pgx.Conn
	pb.UnimplementedExampleServer
}

func (s *server) run() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serv := grpc.NewServer()
	pb.RegisterExampleServer(serv, s)
	log.Printf("Server listening at %v", lis.Addr())

	return serv.Serve(lis)
}

func (s *server) GetById(ctx context.Context, in *pb.QueryId) (*pb.Game, error) {
	gameId := in.GetId()
	log.Printf("Querying game with ID: %d", gameId)

	var game pb.Game
	query := "SELECT appid, name, price FROM steam_games WHERE appid = $1"
	err := dbpool.QueryRow(ctx, query, gameId).Scan(
		&game.Id,
		&game.Name,
		&game.Price,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Juego con ID %d no encontrado", gameId)
		}
		return nil, err
	}

	return &game, nil
}

func (s *server) GetByName(ctx context.Context, in *pb.QueryName) (*pb.Game, error) {
	return nil, nil
}

func (s *server) GetAll(ctx context.Context, empty *pb.Empty) (*pb.GamesList, error) {
	return nil, nil
}

func main() {
	database_url := "postgresql://postgres:Araya123@localhost:5432/steam"
	var err error
	dbpool, err = pgxpool.New(context.Background(), database_url)
	if err != nil {
		log.Fatalf("No se pudo establecer conexión con PostgreSQL: %v", err)
	}
	var example_server *server = NewexampleServer()
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v", err)
	}

	defer conn.Close(context.Background())
	example_server.conn = conn
	if err := example_server.run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
