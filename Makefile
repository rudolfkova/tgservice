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

.PHONY: mocks
mocks:
	mockery --name=SessionUsecase --dir=./internal/port/handler --output=./mocks --outpkg=mocks
	mockery --name=MessageUsecase --dir=./internal/port/handler --output=./mocks --outpkg=mocks
	mockery --name=TelegramMessenger --dir=./internal/usecase --output=./mocks --outpkg=mocks
	mockery --name=TelegramDriver --dir=./internal/usecase --output=./mocks --outpkg=mocks
