.PHONY: up down logs test integration fmt vet openapi smoke
up:
	docker compose up --build -d --wait

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

test:
	cd services && go test ./...

integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo 'Set TEST_DATABASE_URL to a dedicated disposable PostgreSQL database' && exit 1)
	cd services && go test -race ./tests/integration -v

fmt:
	cd services && gofmt -w .

vet:
	cd services && go vet ./...

openapi:
	python3 scripts/generate-openapi.py

smoke:
	python3 scripts/smoke.py
