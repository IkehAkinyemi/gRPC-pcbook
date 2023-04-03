package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/IkehAkinyemi/pcbook/pb"
	"github.com/IkehAkinyemi/pcbook/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey     = "secret-xxxx"
	tokenDuration = 15 * time.Minute
)

const (
	serverCertFile = "cert/server-cert.pem"
	serverKeyFile = "cert/server-key.pem"
	clientCACertFile = "cert/ca-cert.pem"
)

func main() {
	port := flag.Int("port", 0, "server port value")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	serverType := flag.String("srv-type", "grpc", "type of server (grpc/rest)")
	endpoint := flag.String("grpc-endpoint", "", "gRPC endpoint")
	flag.Parse()

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

	address := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		 log.Fatalf("cannot connect tcp listener: %v", err)
	}

	if *serverType == "grpc" {
		err = runGRPCServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
	} else {
		err = runRESTServer(authServer, laptopServer, jwtManager, *enableTLS, listener, *endpoint)
	}

	if err != nil {
		log.Fatal("cannot not start server: %w", err)
	}
}

func runGRPCServer(
	authServer *service.AuthServer,
	laptopServer *service.LaptopServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOption := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %w", err)
		}
		serverOption = append(serverOption, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(serverOption...)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	log.Printf("Start gRPC server at %s, with TLS = %t", listener.Addr().String(), enableTLS)

	return grpcServer.Serve(listener) 
}

func runRESTServer(
	authServer *service.AuthServer,
	laptopServer *service.LaptopServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
	grpcEndpoint string,
) error {
	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// in-process handler
	// err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return err
	}

	// err = pb.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	err = pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts)
	if err != nil {
		return err
	}

	log.Printf("Start REST server at %s, with TLS = %t", listener.Addr().String(), enableTLS)
	if enableTLS {
		return http.ServeTLS(listener, mux, serverCertFile, serverKeyFile)
	}

	return http.Serve(listener, mux)
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// load certificate of the CA who signed client's certificate.
	pemClientCA, err := os.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
