PORT=8080
gen:
	protoc -I=proto --go_out=. --go-grpc_out=. proto/*.proto

clean:
	rm -rf pb/*.go

server:
	go run cmd/server/main.go -port $(PORT)

client:
	go run cmd/client/main.go -srv-addr "localhost:$(PORT)"

test:
	go test -cover -race ./...