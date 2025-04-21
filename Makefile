.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

-include .env
MIGRATIONS_DIR=./db/migrations

migrate:
	migrate -path "$(MIGRATIONS_DIR)" -database $(DATABASE_URI) $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

migration:
	migrate create -ext sql -seq -digits 3 -dir "$(MIGRATIONS_DIR)" $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

coverprofile:
	go test ./... -coverprofile cover.out.tmp && cat cover.out.tmp | grep -v -e "mock_" -e "logging.go" -e ".sql.go" -e "db.go" -e "test" > cover.out && rm cover.out.tmp && go tool cover -func cover.out

ifneq ($(filter $(MAKECMDGOALS),migrate migration covarage),)
%:
	@true
endif
