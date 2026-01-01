.PHONY: up down proto

# Запустить инфраструктуру (БД)
up:
	docker-compose up -d

# Остановить инфраструктуру
down:
	docker-compose down

# Генерация Go-кода из Proto-файлов (понадобится позже)
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/order/order.proto