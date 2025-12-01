BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2

lint: install-lint-deps
	golangci-lint run ./...

migrate:
	go run ./cmd/migrate -dsn "${DB_DSN}" -migrations ./migrations

run-postgres:
	docker-compose -f deployments/docker-compose.yaml up -d db

test-integration:
	# запуск интеграционных тестов в контейнере
	go test -v ./... -run Integration

test-storage:
	# тестирование хранилища с добавлением данных
	go run ./cmd/test_storage -config ./configs/config.yaml

generate:
	# генерация GRPC кода из proto файлов
	@which protoc > /dev/null || (echo "protoc not found. Install: https://grpc.io/docs/protoc-installation/" && exit 1)
	@which protoc-gen-go > /dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@which protoc-gen-go-grpc > /dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	mkdir -p api/event
	protoc --go_out=api/event --go_opt=paths=source_relative \
		--go-grpc_out=api/event --go-grpc_opt=paths=source_relative \
		--proto_path=api api/EventService.proto

.PHONY: build run build-img run-img version test lint test-storage generate