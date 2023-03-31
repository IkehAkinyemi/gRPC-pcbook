package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/IkehAkinyemi/pcbook/pb"
	"github.com/IkehAkinyemi/pcbook/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey     = "secret-xxxx"
	tokenDuration = 15 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "server port value")
	flag.Parse()
	log.Printf("starting server on port: %d", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	privateKey, publicKey, err := readInConfig("./keys/")
	if err != nil {
		log.Fatal(err)
	}

	jwtManager := service.NewJWTManager(privateKey, publicKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

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

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}
	return createUser(userStore, "user1", "secret", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	laptopServerPath := "/LaptopService/"

	return map[string][]string{
		laptopServerPath + "CreateLaptop": {"admin"},
		laptopServerPath + "RateLaptop":   {"admin", "user"},
		laptopServerPath + "UploadImage":  {"admin"},
	}
}

func readInConfig(path string) (string, string, error) {
	privateKey, err := os.ReadFile(path + "private_key.pem")
	if err != nil {
		return "", "", err
	}

	publicKey, err := os.ReadFile(path + "public_key.pem")
	if err != nil {
		return "", "", err
	}

	return string(privateKey), string(publicKey), nil
}
