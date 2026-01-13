APP_BIN := "./bin/calendar"
SCHEDULER_BIN := "./bin/calendar_scheduler"
SENDER_BIN := "./bin/calendar_sender"

DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --pretty="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

DOCKER_COMPOSE_INTEGRATION=docker-compose --file=deployments/docker-compose.integration.yaml --env-file=testing.env

build:
	go build -v -o $(APP_BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/scheduler
	go build -v -o $(SENDER_BIN) -ldflags "$(LDFLAGS)" ./cmd/sender

run: build
	$(APP_BIN) -config ./configs/calendar/config.yaml
	$(SCHEDULER_BIN) -config ./configs/scheduler/config.yaml
	$(SENDER_BIN) -config ./configs/sender/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(APP_BIN) version

test:
	go test -skip TestIntegration -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.3.0

lint: install-lint-deps
	golangci-lint run ./... --config=./.golangci.yml

migrate:
	goose up

generate:
	rm -rf internal/pb
	mkdir -p internal/pb

	protoc --proto_path=api/calendar/ --go_out=internal/pb --go-grpc_out=internal/pb api/calendar/*.proto

rebuild-containers:
	docker-compose build --no-cache

up:
	docker-compose --file=./deployments/docker-compose.yaml up -d

down:
	docker-compose --file=./deployments/docker-compose.yaml down

restart:
	make down
	make up

rebuild-containers-testing:
	$(DOCKER_COMPOSE_INTEGRATION) --profile=tests build --no-cache

up-testing:
	$(DOCKER_COMPOSE_INTEGRATION) up -d

down-testing:
	$(DOCKER_COMPOSE_INTEGRATION) down

integration-tests:
	$(DOCKER_COMPOSE_INTEGRATION) up -d --wait

	$(DOCKER_COMPOSE_INTEGRATION) run --rm -P integration-tests -test.v

	TEST_EXIT_CODE=$$?

	make down-testing

	exit $$TEST_EXIT_CODE

.PHONY: build run build-img run-img version test lint
