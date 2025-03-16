swag:
	swag init -g ./cmd/main/main.go

run:
	docker-compose build && docker-compose up -d
