LOCAL_BIN:=$(CURDIR)/bin

GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
GOLANGCI_TAG:=1.62.2
MOCKGEN_TAG:=latest
SMART_IMPORTS := ${LOCAL_BIN}/smartimports

# run in docker
.PHONY: docker
docker:
	godotenv -f ./bin/dev.env docker-compose up --build -d

# build app
.PHONY: build
build:
	go mod download && CGO_ENABLED=0  go build \
		-o ./bin/main$(shell go env GOEXE) ./cmd/server/main.go


# run dev
.PHONY: dev
dev:
	godotenv -f ./bin/dev.env go run ./cmd/server/main.go

.PHONY: db
db:
	godotenv -f ./bin/dev.env docker compose up db 

# install golangci-lint binary
.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info Downloading golangci-lint v$(GOLANGCI_TAG))
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
endif

# run diff lint like in pipeline
.PHONY: .lint
.lint: install-lint
	$(info Running lint...
	$(GOLANGCI_BIN) run --new-from-rev=origin/main --config=.golangci.yaml ./...

# golangci-lint diff main
.PHONY: lint
lint: .lint

# run full lint like in pipeline
.PHONY: .lint-full
.lint-full: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yaml ./...

# golangci-lint full
.PHONY: lint-full
lint-full: .lint-full

.PHONY: format
format:
	$(info Running goimports...)
	test -f ${SMART_IMPORTS} || GOBIN=${LOCAL_BIN} go install github.com/pav5000/smartimports/cmd/smartimports@latest
	${SMART_IMPORTS} -exclude pkg/,internal/pb  -local 'github.com'


.PHONY: coverage
coverage:
	go test -coverprofile bin/cover.out ./internal/...
	go tool cover -html=bin/cover.out


.PHONY: migration-create
migration-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migration-create NAME=your_migration_name"; \
		exit 1; \
	fi
	goose -dir migrations create $(NAME) sql


