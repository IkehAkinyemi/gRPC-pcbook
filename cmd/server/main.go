package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/IkehAkinyemi/pcbook/pb"
	"github.com/IkehAkinyemi/pcbook/service"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 0, "server port value")
	flag.Parse()
	log.Printf("starting server on port: %d", *port)

	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("cannot connect tcp listener: %v", err)
	}

	fmt.Printf("%s\n", listener.Addr().String())

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("cannot not start server: %v", err)
	}
}