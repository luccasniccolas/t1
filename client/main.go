package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/luccasniccolas/t1/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewExampleClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var id int32 = 12100
	queryId := &pb.QueryId{Id: id}
	r, err := c.GetById(ctx, queryId)
	if err != nil {
		log.Fatalf("No encontro el la query: %v", err)
	}

	fmt.Printf("r.GetById(): %v", r.Name)
}
