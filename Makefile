.PHONY: docker-build run migrate clean

docker-build:
	docker build -t avito-pvz-service ..

docker-down:
	docker-compose down -v
	docker-compose down --rmi all

run:
	docker-compose up --build

test:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html:
	go tool cover -html=coverage.out -o coverage.html
	start coverage.html

clean:
	del /Q /F coverage
	del /Q /F *.out
	del /Q /F *.html

migrate:
	psql -h localhost -U avito -d avito_db -f migrations/0001_create_tables.sql
