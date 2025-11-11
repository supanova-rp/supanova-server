GIT_HASH := $(shell git rev-parse --short HEAD)
DOCKER_USER := jdgarner
IMAGE_NAME := go-template

dep:
	go mod download

run:
	go run main.go

lint: lint/install lint/run

lint/install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.5.0

lint/run:
	bin/golangci-lint run --config .golangci.yml

test:
	go test ./...

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate -f internal/store/sqlc.yaml

migrate/create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate/create name=<migration_name>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir internal/store/migrations -seq $(name)

build:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o go-template .

docker/local-build:
	DOCKER_BUILDKIT=1 docker buildx build \
	-t $(DOCKER_USER)/$(IMAGE_NAME):local .

docker/ci-build:
	DOCKER_BUILDKIT=1 docker buildx build \
	--platform linux/amd64 \
	-t $(DOCKER_USER)/$(IMAGE_NAME):latest \
	-t $(DOCKER_USER)/$(IMAGE_NAME):$(GIT_HASH) .
