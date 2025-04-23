test:
	go test -count=1 ./...

generate:
	go generate

build: generate
	go build -o ./bin/chat main.go

run: build mongo-up
	./bin/chat

install:
	go install github.com/swaggo/swag/cmd/swag@latest

mongo-up:
	docker compose up -d

mongo-down:
	docker compose down

fmt:
	go fmt ./...
	swag fmt
