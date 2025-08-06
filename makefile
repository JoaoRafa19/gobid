migrate:
	go run cmd/terndotenv/main.go

gen:
	go generate ./...

run:
	go run ./cmd/api/main.go