package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/luccasniccolas/t1/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Establece la conexión con Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "172.20.0.7:6379",
		Password: "replica1234",
		DB:       0,
	})

	// Verifica la conexión con Redis
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Error al conectar a Redis:", err)
		return
	}
	fmt.Println("Conexión exitosa a Redis:", pong)

	reader := bufio.NewReader(os.Stdin) // Para leer desde la consola

	for {
		// Lee el ID desde la consola
		fmt.Print("Ingrese un ID (-1 para GetAll, 0 para salir): ")
		input, _ := reader.ReadString('\n')
		var id int32
		_, err = fmt.Sscan(input, &id)
		if err != nil {
			fmt.Println("Error al leer el ID:", err)
			continue
		}

		// Si el ID es 0, se sale del bucle
		if id == 0 {
			fmt.Println("Saliendo...")
			break
		}

		// Empieza el proceso para el ID ingresado
		start := time.Now()
		if id == -1 {
			// Llama a GetAll en lugar de GetById
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("No se pudo conectar: %v", err)
			}
			defer conn.Close()
			c := pb.NewExampleClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			empty := &pb.Empty{}
			gamesList, err := c.GetAll(ctx, empty)
			if err != nil {
				log.Fatalf("No se pudieron obtener todos los juegos: %v", err)
			}

			for _, game := range gamesList.List {
				fmt.Printf("Resultado de GetAll(): %v %v %v \n", game.Id, game.Name, game.Price)
			}

		} else {
			// Llama a GetById con el ID especificado
			resultJSON, err := client.Get(context.Background(), fmt.Sprintf("result:%d", id)).Result()
			if err == nil {
				// Si el resultado está en la caché, lo decodificamos desde JSON
				var r pb.Game
				err := json.Unmarshal([]byte(resultJSON), &r)
				if err != nil {
					log.Fatalf("Error al decodificar el resultado desde JSON: %v", err)
				}
				fmt.Printf("Resultado obtenido desde la caché Redis: %v %v %v \n", r.Id, r.Name, r.Price)
			} else if err != redis.Nil {
				log.Fatalf("Error al obtener el resultado desde Redis: %v", err)
			} else {
				// Si el resultado no está en la caché, realiza la consulta al servidor gRPC
				conn, err := grpc.Dial(address, grpc.WithInsecure())
				if err != nil {
					log.Fatalf("No se pudo conectar: %v", err)
				}
				defer conn.Close()
				c := pb.NewExampleClient(conn)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				queryId := &pb.QueryId{Id: id}
				r, err := c.GetById(ctx, queryId)
				if err != nil {
					log.Fatalf("No se encontró la query: %v", err)
				}

				fmt.Printf("Resultado de GetById(): %v %v %v \n", r.Id, r.Name, r.Price)

				// Convierte el resultado a JSON y lo almacena en la caché Redis
				resultJSON, err := json.Marshal(r)
				if err != nil {
					log.Fatalf("Error al convertir el resultado a JSON: %v", err)
				}
				err = client.Set(ctx, fmt.Sprintf("result:%d", id), resultJSON, 60*time.Second).Err()
				if err != nil {
					log.Fatalf("Error al almacenar en Redis: %v", err)
				}
				fmt.Println("Resultado almacenado en Redis con éxito.")
			}
		}

		elapsed := time.Since(start)
		fmt.Printf("Tiempo de consulta total: %v\n", elapsed)
	}
}
