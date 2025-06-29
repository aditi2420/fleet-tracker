.PHONY: build up down logs token test test-verbose test-coverage test-unit test-integration test-short

build:           ## build API image only
	docker compose build api

up:              ## build + run entire stack
	docker compose up -d --build

down:            ## stop & remove containers + volumes
	docker compose down -v

logs:            ## follow API logs
	docker compose logs -f api

token:           ## generate a dev JWT
	go run ./cmd/token -sub dev -exp 24h

test:            ## run all tests
	go test ./...

