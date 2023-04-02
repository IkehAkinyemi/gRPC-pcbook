PORT=8080
gen:
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm -rf pb/*.go

server:
	go run cmd/server/main.go -port $(PORT)

client:
	go run cmd/client/main.go -srv-addr "localhost:$(PORT)"

keys:
	openssl genpkey -algorithm RSA -out keys/private_key.pem
	openssl rsa -in keys/private_key.pem -out keys/public_key.pem -pubout

test:
	go test -cover -race ./...

cert:
	cd cert; ./generate_ssl_cert.sh; cd ..

.PHONY: gen clean server client keys test cert