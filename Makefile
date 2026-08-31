.PHONY: all build clean test docker-build docker-up docker-down

BIN_DIR=bin
GO_BUILD=CGO_ENABLED=0 go build -ldflags="-s -w"

all: build

build:
	mkdir -p $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/aegis-gateway ./cmd/aegis-gateway
	$(GO_BUILD) -o $(BIN_DIR)/merchant-mcp ./cmd/merchant-mcp
	$(GO_BUILD) -o $(BIN_DIR)/dashboard ./cmd/dashboard
	$(GO_BUILD) -o $(BIN_DIR)/demo-buyer ./cmd/demo-buyer
	$(GO_BUILD) -o $(BIN_DIR)/redteam ./cmd/redteam
	$(GO_BUILD) -o $(BIN_DIR)/seed-catalog ./cmd/seed-catalog
	$(GO_BUILD) -o $(BIN_DIR)/migrate ./cmd/migrate
	$(GO_BUILD) -o $(BIN_DIR)/gateway ./cmd/gateway

clean:
	rm -rf $(BIN_DIR)

test:
	go test -v ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
