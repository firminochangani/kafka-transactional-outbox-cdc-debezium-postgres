include .env
export

install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

setup-debezium:
	./config/setup-debezium.sh

setup-kafka:
	./config/setup-kafka.sh

mig-up:
	goose -dir config/sql up

mig-down:
	goose -dir config/sql down

order-svc:
	reflex -s -r '\.go$$' go run ./ order-svc

inventory-svc:
	reflex -s -r '\.go$$' go run ./ inventory-svc

