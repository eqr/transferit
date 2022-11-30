# replace app with actual app name
.PHONY: test
test:
	go test ./...


.PHONY: dockerize
dockerize:
	docker build -t app .


.PHONY: run-dockerized
run-dockerized:
	docker run -p 8080:8080 app


.PHONY: start
start:
	docker-compose build
	docker network create app-network
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose down
	docker network rm app-network

build:
	docker-compose build

run:
	docker-compose up -d

deploy: build run

connect:
	docker exec -it app /bin/sh
