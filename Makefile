.PHONY: docker-build run migrate

docker-build:
	docker build -t avito-pvz-service ..

docker-down:
	docker-compose down -v

run:
	docker-compose up --build

migrate:
	psql -h localhost -U avito -d avito_db -f migrations/0001_create_tables.sql
