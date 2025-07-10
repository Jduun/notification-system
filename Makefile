include .env
export

env:
	cp .env.example .env

clean:
	sudo rm -rf $$DB_PATH $$KAFKA_PATH
	make compose-down

down:
	docker compose down --remove-orphans

build:
	docker compose up --build -d && docker compose logs -f

up:
	docker compose up -d && docker compose logs -f

stop:
	docker compose stop

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt