PORT=8080
gen:
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm -rf pb/*.go

server1:
	go run cmd/server/main.go -port 50051

server2:
	go run cmd/server/main.go -port 50052

server1-tls:
	go run cmd/server/main.go -port 50051 -tls

server2-tls:
	go run cmd/server/main.go -port 50052 -tls

server:
	go run cmd/server/main.go -port $(PORT)

client-tls:
	go run cmd/client/main.go -srv-addr "localhost:$(PORT)" -tls

client:
	go run cmd/client/main.go -srv-addr "localhost:$(PORT)"

keys:
	openssl genpkey -algorithm RSA -out keys/private_key.pem
	openssl rsa -in keys/private_key.pem -out keys/public_key.pem -pubout

test:
	go test -cover -race ./...

cert:
	cd cert; ./generate_ssl_cert.sh; cd ..

.PHONY: gen clean server client keys test cert server1 server2 server1-tls server2-tls client-tls