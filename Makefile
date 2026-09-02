.PHONY: help up down build test test-backend test-frontend clean

help:
	@echo "Comandos disponíveis:"
	@echo "  make up            - Inicia todos os serviços via Docker Compose"
	@echo "  make down          - Para todos os serviços"
	@echo "  make build         - Constrói as imagens Docker"
	@echo "  make test          - Roda todos os testes (Go + React)"
	@echo "  make test-backend  - Roda os testes unitários do Backend Go (TDD)"
	@echo "  make test-frontend - Roda os testes unitários do Frontend React"

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

test: test-backend test-frontend

test-backend:
	cd backend && go test -v -race ./...

test-frontend:
	cd frontend && npm test -- --run

clean:
	docker compose down -v
