package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/IkehAkinyemi/pcbook/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopClient is a client to call laptop service RPC.
type LaptopClient struct {
	service pb.LaptopServiceClient
}

// NewLaptopClient return an instance of LaptopClient.
func NewLaptopClient(conn *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(conn)
	return &LaptopClient{service}
}

// CreateLaptop send a create laptop request.
func (client LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Println("laptop already exist")
		} else {
			log.Fatal("cannot create laptop", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}

// SearchLaptop sends a search laptop request using filter params.
func (client LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Printf("search filter: %+v", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.service.SearchLaptop(ctx, req)
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

// UploadImage uploads a laptop picture to the server.
func (client LaptopClient) UploadImage(laptopID, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.UploadImage(ctx)
	if err != nil {
		log.Fatalf("cannot upload image: %v", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1<<10)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("csnnot read chunk to buffer: %v", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

// RateLaptop send a laptop rating request.
func (client LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitReponse := make(chan error)

	// go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more responses")
				waitReponse <- nil
				return
			}
			if err != nil {
				waitReponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}
			log.Print("received response: ", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("sent request: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitReponse
	return err
}
