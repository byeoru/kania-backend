server:
	go run init/main.go

postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=roalqufn-12 -d postgres:17-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root kania

dropdb:
	docker exec -it postgres dropdb kania

migrateup:
	migrate -path db/migration -database "postgresql://root:roalqufn-12@localhost:5432/kania?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:roalqufn-12@localhost:5432/kania?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:roalqufn-12@localhost:5432/kania?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:roalqufn-12@localhost:5432/kania?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

.PHONY: server postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc