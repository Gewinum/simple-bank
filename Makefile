ifndef CONTAINER
override CONTAINER = bank-postgres
endif

postgres:
	docker run --name $(CONTAINER) -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -p 5432:5432 -d postgres:17-alpine

createdb:
	docker exec -it $(CONTAINER) createdb -U root simplebank

dropdb:
	docker exec -it $(CONTAINER) dropdb -U root simplebank

migrateup:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simplebank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simplebank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test
