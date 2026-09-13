TEST_COMPOSE_PROJECT ?= vendor-onboarding-tracker-test

.PHONY: up down down-volumes logs migrate seed seed-reset test test-unit test-integration

up:
	docker compose up --build

down:
	docker compose down

down-volumes:
	docker compose down --volumes --remove-orphans

logs:
	docker compose logs --follow

migrate:
	docker compose run --rm migrate

seed:
	docker compose run --rm seed

seed-reset:
	docker compose --profile tools run --rm seed-reset

test: test-unit

test-unit:
	go test ./...
	pnpm --dir web test --run

test-integration:
	@status=0; \
		docker compose --project-name $(TEST_COMPOSE_PROJECT) --profile test up --detach postgres-test migrate-test || status=$$?; \
		if [ $$status -eq 0 ]; then \
			TEST_DATABASE_URL="postgres://onboarding:onboarding@localhost:$${POSTGRES_TEST_PORT:-5435}/vendor_onboarding_test?sslmode=disable" go test -tags=integration -p 1 ./... || status=$$?; \
		fi; \
		docker compose --project-name $(TEST_COMPOSE_PROJECT) --profile test down --volumes; \
		exit $$status
