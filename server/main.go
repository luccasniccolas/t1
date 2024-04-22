package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

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
			return nil, fmt.Errorf("Game ID %d not Found", gameId)
		}
		return nil, err
	}

	return &game, nil
}

func (s *server) GetByName(ctx context.Context, in *pb.QueryName) (*pb.Game, error) {
	return nil, nil
}

func (s *server) GetAll(ctx context.Context, empty *pb.Empty) (*pb.GamesList, error) {
	log.Println("Realizando consulta para obtener todos los juegos")
	// Consulta para obtener todos los registros
	query := "SELECT appid, name, release_date, required_age, categories, price FROM steam_games"

	// Ejecuta la consulta
	rows, err := dbpool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta SQL: %v", err)
	}
	defer rows.Close()

	// Inicializa la lista de juegos para devolver
	gamesList := &pb.GamesList{
		List: make([]*pb.Game, 0), // Inicializa con un slice vacío de juegos
	}

	// Itera sobre los resultados y construye la lista de juegos
	for rows.Next() {
		var game pb.Game
		var releaseDate time.Time
		err := rows.Scan(
			&game.Id,
			&game.Name,
			&releaseDate,
			&game.RequiredAge,
			&game.Categories,
			&game.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear la fila: %v", err)
		}

		// Agrega el juego a la lista
		gamesList.List = append(gamesList.List, &game)
	}

	// Comprueba si hubo errores en el proceso de iteración
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante el recorrido de las filas: %v", err)
	}

	return gamesList, nil
}

func main() {
	database_url := "postgresql://user:user@172.20.0.5:5432/tarea1"
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
