include .env

dev:
	go run ./cmd/drunk/main.go

create_migration:
	goose -dir sql/schema create $(name) sql

up_by_one:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE) goose -dir sql/schema up-by-one

upse:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE) goose -dir sql/schema up

downse:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE) goose -dir sql/schema down 

resetse:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE) goose -dir sql/schema reset

sql_gen:
	sqlc generate

swag:
	swag init -g ./cmd/drunk/main.go -o ./docs --parseDependency --parseInternal

.PHONY: dev downse upse resetse up_by_one create_migration sql_gen swag

.PHONY: air
