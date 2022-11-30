FROM golang:1.19-alpine

WORKDIR /usr/local/go/src/app/

COPY vendor/ ./vendor
COPY go.mod  ./
COPY go.sum  ./

# combine everything in build directory
RUN mkdir build
COPY app/templates/ ./build/templates
COPY app/run/docker.yml ./build/docker.yml

COPY app/run/docker.yml ./build

# build steps
COPY app/ ./app
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/app ./app/run

EXPOSE 8082

WORKDIR /usr/local/go/src/app/build
ENV GIN_MODE=release
ENTRYPOINT ["./app", "--config", "docker.yml"]
