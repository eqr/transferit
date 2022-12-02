# replace app with actual app name
.PHONY: test
test:
	go test ./...

.PHONY: dockerize
dockerize:
	docker build -t trasferit .

.PHONY: run-dockerized
run-dockerized:
	docker run -p 8080:8080 transferit

.PHONY: start
start:
	docker-compose build
	docker network create transferit-network
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose down
	docker network rm transferit-network

build:
	docker-compose build

run:
	docker-compose up -d

deploy: build run

connect:
	docker exec -it transferit /bin/sh
