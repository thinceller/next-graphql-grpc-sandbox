DB_PASSWORD := "password"

.PHONY: mysql
mysql:
	docker-compose exec db mysql -uroot -p${DB_PASSWORD} todo

.PHONY: migrate/status
migrate/status:
	goose -dir ./migrate mysql "root:password@tcp(127.0.0.1:9000)/todo?parseTime=true" status

.PHONY: migrate/up
migrate/up:
	goose -dir ./migrate mysql "root:password@tcp(127.0.0.1:9000)/todo?parseTime=true" up

.PHONY: clean
clean:
	rm -rf out/*
	go clean

.PHONY: build
build: clean
	go build -o out/todo-app .
