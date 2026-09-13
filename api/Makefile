include .env
export

MIGRATIONS_DIR := db/migrations
DB_DSN := postgres://$(DATABASE_USERNAME):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)$(APP_ENV)?sslmode=disable

MIGRATE := migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)"

.PHONY: migrate-install migrate-new migrate-up migrate-down migrate-down-all migrate-force migrate-version schema-dump

## Install the golang-migrate CLI (postgres driver only).
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## Create a new migration pair: make migrate-new NAME=add_something
migrate-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## Apply all pending migrations.
migrate-up:
	$(MIGRATE) up

## Roll back the last applied migration.
migrate-down:
	$(MIGRATE) down 1

## Roll back every migration.
migrate-down-all:
	$(MIGRATE) down -all

## Force the schema_migrations version without running any migration: make migrate-force VERSION=1
migrate-force:
	$(MIGRATE) force $(VERSION)

## Print the current migration version.
migrate-version:
	$(MIGRATE) version

## Dump the current (post-migration) schema to db/schema.sql. Regenerate
## this after adding a migration -- it is a generated snapshot for reading
## and for sqlc, not a second source of truth.
schema-dump:
	docker compose exec -T foodstore_pgdb pg_dump --schema-only --no-owner --no-privileges --no-comments \
		-U $(DATABASE_USERNAME) -d $(DATABASE_NAME)$(APP_ENV) \
		-T schema_migrations > db/schema.sql
