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
	go test ./... -covermode=count -coverprofile cover.out.tmp && cat cover.out.tmp | grep -v -e "mocks" -e "test" -e "logging" > cover.out \
 		&& rm cover.out.tmp && go tool cover -html cover.out -o coverprofile.html && go tool cover -func cover.out

ifneq ($(filter $(MAKECMDGOALS),migrate migration covarage),)
%:
	@true
endif
