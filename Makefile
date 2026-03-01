.PHONY: gen
gen:
	protoc -I . \
	  --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  proto/tgservice/v1/tgservice.proto

.PHONY: build
build:
	go build -v ./cmd/app

.PHONY: start
start:
	./app.exe