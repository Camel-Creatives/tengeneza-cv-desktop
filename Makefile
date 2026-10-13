.PHONY: run worker test tidy build build-api frontend desktop dev lint docker-up

run:
	go run ./cmd/api

worker:
	go run ./cmd/worker

test:
	go test ./...

tidy:
	go mod tidy

build: build-api

build-api:
	go build ./cmd/api ./cmd/worker

frontend:
	pnpm --dir frontend vite:build

desktop: frontend
	mkdir -p build
	go build -o build/TengenezaCV .

dev:
	wails dev

lint:
	go vet ./...

docker-up:
	docker compose up -d postgres redis minio
