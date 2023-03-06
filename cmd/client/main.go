package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/IkehAkinyemi/pcbook/pb"
	"github.com/IkehAkinyemi/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	address := flag.String("srv-addr", "", "server address to dial")
	flag.Parse()
	log.Printf("dialing server %s", *address)

	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error occurred dialing address: %v", err)
	}

	client := pb.NewLaptopServiceClient(conn)

	laptop := sample.NewLaptop()
	laptop.Id = "invalid"

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	res, err := client.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Println("laptop already exist")
		} else {
			log.Fatal("cannot create laptop")
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}