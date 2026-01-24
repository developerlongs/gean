.PHONY: build test clean run help generate lint

BIN_DIR := bin
BINARY := $(BIN_DIR)/gean

# Configuration
GENESIS_DIR ?= 
NODE_ID ?= 
LISTEN_ADDR ?= /ip4/0.0.0.0/tcp/9000
LOG_LEVEL ?= info

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate: ## Generate SSZ encoding code
	go run github.com/ferranbt/fastssz/sszgen --path=./consensus --objs=Checkpoint,Config,Vote,SignedVote,BlockHeader,BlockBody,Block,SignedBlock,State

build: ## Build the gean binary
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/gean

test: ## Run tests
	go test ./... -v

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
	go clean

lint: ## Run go vet
	go vet ./...

run: ## Run gean (requires GENESIS_DIR)
ifndef GENESIS_DIR
	@echo "Usage: make run GENESIS_DIR=/path/to/genesis [NODE_ID=gean_0]"
	@exit 1
else
	@$(MAKE) build
ifdef NODE_ID
	./$(BINARY) $(GENESIS_DIR) --node-id $(NODE_ID) --listen "$(LISTEN_ADDR)" --log-level $(LOG_LEVEL)
else
	./$(BINARY) $(GENESIS_DIR) --listen "$(LISTEN_ADDR)" --log-level $(LOG_LEVEL)
endif
endif
