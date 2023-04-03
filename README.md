# PCBook Inventory
This project implements gRPC (unary, server-streaming, client-streaming and bidirectional-streaming) and REST API servers and clients using Go, Protocol Buffers, gRPC, and the gRPC Gateway library.

## Summary
The project uses the `protoc` command to generate Go code for gRPC server and client, as well as an OpenAPI specification for the REST API. The `openssl` command is also used to generate private and public keys for TLS support and JWT RSA keys.

The project consists of two gRPC servers, server1 and server2, which can be run on ports 50051 and 50052, respectively. Both servers can be started with or without TLS support using the server1, server2, server1-tls, and server2-tls targets.

In addition to the gRPC servers, the project includes a REST API server that can be started using the rest target. The REST API server listens on port 8081 and forwards requests to the gRPC server running on port 8080.

The project also includes a client that can connect to the gRPC server running on the specified address and port. The client can be started with or without TLS support using the client and client-tls targets.

## Requirements
To build and run this project, you will need:

- Go v1.16 or later
- Protoc (libprotoc 3.21.12) or later
- gRPC Gateway (protoc-gen-grpc-gateway) v2.1.0 or later
- gRPC OpenAPI (protoc-gen-openapiv2)
- OpenSSL
- Nginx

# Setup
Clone the repository:
```bash
git clone git@github.com:IkehAkinyemi/gRPC-pcbook.git
```
Generate SSL certificates:
```sh
make cert
```
Generate Go code from the Protocol Buffer files:
```sh
make gen
```

## Usage
### Running the gRPC servers
There are two test gRPC servers included in this project for testing the use of `nginx` for load balancing:

- server1: listens on port 50051
- server2: listens on port 50052
To start a server, run one of the following commands:

```sh
make server1
make server2
```

### Running the gRPC servers with TLS
To run a gRPC server with TLS, use the following commands:

```sh
make server1-tls

make server2-tls
```

### Running the REST API server
To run the REST API server, use the following command:
```sh
make rest
```
The REST API server listens on port 8081 and proxies requests to the gRPC server running on port 8080.

### Running the gRPC client
To run the gRPC client, use the following command:

```sh
make client
```
By default, the client connects to the gRPC server running on `http://localhost:8080`.

### Running the gRPC client with TLS
To run the gRPC client with TLS, use the following command:

```sh
make client-tls
```
By default, the client connects to the gRPC server running on `http://localhost:8080`. 

> Note that if `nginx` is running and proxing request to the servers on ports 50051 and 50052, then the client connection is with the nginx on port 8080 .

### Generating Private and public keys for JWT
To generate private and public keys, use the following command:

```sh
make keys
```
This will generate a private key and public keys to signing and generating secured JWT tokens for authentication and authorization in the keys directory.

### Running tests
To run tests, use the following command:

```sh
make test
```

### Cleaning up generated files
To clean up the generated Go code, use the following command:
```sh
make clean
```
## License
This project is licensed under the MIT License - see the LICENSE file for details.

For more information on how to use the project, see the documentation in the code files or contact the project maintainers.