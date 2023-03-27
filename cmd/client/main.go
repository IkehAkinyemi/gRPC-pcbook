package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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

	for i := 0; i < 10; i++ {
		createLaptop(client)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit: pb.Memory_GIGABYTE,
		},
	}

	searchLaptop(client, filter)
}

func createLaptop(client pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()

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

func searchLaptop(client pb.LaptopServiceClient, filter *pb.Filter) {
	log.Printf("search filter: %+v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal()
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("cannot receive response: %v", err)
		}
		laptop := res.GetLaptop()
		fmt.Println("– found: ", laptop.GetId())
		fmt.Println("  + brand: ", laptop.GetBrand())
		fmt.Println("  + name: ", laptop.GetName())
		fmt.Println("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
		fmt.Println("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		fmt.Println("  + ram: ", laptop.GetRam(), laptop.GetRam().GetUnit())
		fmt.Println("  + price: ", laptop.GetPriceUsd(), "USD")
	}
}