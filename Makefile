.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

-include .env
MIGRATIONS_DIR=./db/migrations

migrate:
	migrate -path "$(MIGRATIONS_DIR)" -database $(DATABASE_URI) $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

migration:
	migrate create -ext sql -seq -digits 3 -dir "$(MIGRATIONS_DIR)" $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

ifneq ($(filter $(MAKECMDGOALS),migrate migration),)
%:
	@true
endif
