NAME  := tekton-hub-proxy
MD_FILES := $(shell find . -type f -regex ".*md"  -not -regex '^./vendor/.*' -not -regex '^./.vale/.*' -not -regex "^./.git/.*" -print)

LDFLAGS := -s -w
FLAGS += -ldflags "$(LDFLAGS)" -buildvcs=true
OUTPUT_DIR = bin
TEST_FLAGS = -v 
COVERAGE_FLAGS = -coverprofile=coverage.out -covermode=atomic 

all: test lint build
FORCE:

.PHONY: vendor
vendor:
	@echo Generating vendor directory
	@go mod tidy && go mod vendor

$(OUTPUT_DIR)/$(NAME)-aarch64-linux: main.go 
	env GOARCH=arm64 GOOS=linux	go build -mod=vendor $(FLAGS)   -o $@ ./$<

test:
	@go test $(TEST_FLAGS) ./... 

.PHONY: html-coverage
html-coverage: ## generate html coverage
	@mkdir -p tmp
	@go test $(COVERAGE_FLAGS) -coverprofile=tmp/c.out ./.../ && go tool cover -html=tmp/c.out

clean:
	@rm -rf $(OUTPUT_DIR)/gosmee

build: clean
	@echo "building."
	@mkdir -p $(OUTPUT_DIR)/
	@go build  $(FLAGS)  -o $(OUTPUT_DIR)/$(NAME) ./cmd/$(NAME)/

# Run the application
run: build
	@echo "Running $(NAME)..."
	./$(OUTPUT_DIR)/$(NAME)

# Run with debug logging
run-debug: build
	@echo "Running $(NAME) with debug logging..."
	./$(OUTPUT_DIR)/$(NAME) -debug

lint: lint-go lint-md

lint-go:
	@echo "linting."
	golangci-lint version
	golangci-lint run ./... --modules-download-mode=vendor

.PHONY: lint-md
lint-md: ${MD_FILES} ## runs markdownlint on all markdown files
	@echo "Linting markdown files..."
	@markdownlint $(MD_FILES)

fmt:
	@go fmt `go list ./... | grep -v /vendor/`

fumpt:
	@gofumpt -e -w -extra ./

