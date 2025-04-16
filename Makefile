.PHONY: docker-build run migrate

docker-build:
	docker build -t avito-pvz-service ..

docker-down:
	docker-compose down -v
	docker-compose down --rmi all

run:
	docker-compose up --build

migrate:
	psql -h localhost -U avito -d avito_db -f migrations/0001_create_tables.sql
