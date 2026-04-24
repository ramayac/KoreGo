# korego Makefile
# -------------------------------------------------------------------
# All Go is built with CGO_ENABLED=0 for scratch-container compatibility.

BINARY     := korego
CMD        := ./cmd/korego
MODULE     := github.com/ramayac/korego
VERSION    ?= 0.1.0
LDFLAGS    := -ldflags "-s -w -X '$(MODULE)/pkg/common.Version=$(VERSION)' \
                              -X 'main.Version=$(VERSION)'"

# Directories tested by the unit-test and coverage targets.
PKG_DIRS   := ./pkg/common/... \
              ./internal/dispatch/... \
              ./pkg/echo/... \
              ./pkg/truefalse/... \
              ./pkg/whoami/... \
              ./pkg/hostname/... \
              ./pkg/uname/... \
              ./pkg/pwd/... \
              ./pkg/printenv/... \
              ./pkg/env/... \
              ./pkg/yes/...

.DEFAULT_GOAL := help

# -------------------------------------------------------------------
# help — list all targets with descriptions
# -------------------------------------------------------------------
.PHONY: help
help:
	@echo ""
	@echo "  korego — $(VERSION)"
	@echo ""
	@echo "  Usage: make <target>"
	@echo ""
	@echo "  Build"
	@echo "    build        Compile the korego binary (CGO_ENABLED=0)"
	@echo "    build-race   Compile with -race detector (dev only)"
	@echo "    install      Install korego to \$$GOPATH/bin"
	@echo ""
	@echo "  Test"
	@echo "    test         Run all unit tests"
	@echo "    test-v       Run all unit tests (verbose)"
	@echo "    cover        Run tests and open HTML coverage report"
	@echo "    cover-pct    Print per-package coverage percentages"
	@echo ""
	@echo "  Quality"
	@echo "    vet          Run go vet"
	@echo "    lint         Run staticcheck (installs if missing)"
	@echo "    fmt          Run gofmt -w on all Go files"
	@echo "    fmt-check    Check formatting without modifying files"
	@echo ""
	@echo "  Smoke"
	@echo "    smoke        Build + run manual integration smoke tests"
	@echo "    symlink-test Test symlink dispatch (ln -s korego echo)"
	@echo ""
	@echo "  Housekeeping"
	@echo "    clean        Remove build artifacts"
	@echo "    tidy         go mod tidy"
	@echo "    all          vet + test + build"
	@echo ""

# -------------------------------------------------------------------
# Build targets
# -------------------------------------------------------------------
.PHONY: build
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(CMD)

.PHONY: build-race
build-race:
	go build -race $(LDFLAGS) -o $(BINARY)-race $(CMD)

.PHONY: install
install:
	CGO_ENABLED=0 go install $(LDFLAGS) $(CMD)

# -------------------------------------------------------------------
# Test targets
# -------------------------------------------------------------------
.PHONY: test
test:
	CGO_ENABLED=0 go test $(PKG_DIRS)

.PHONY: test-v
test-v:
	CGO_ENABLED=0 go test -v $(PKG_DIRS)

.PHONY: cover
cover:
	CGO_ENABLED=0 go test -coverprofile=coverage.out $(PKG_DIRS)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"
	@command -v xdg-open >/dev/null 2>&1 && xdg-open coverage.html || true

.PHONY: cover-pct
cover-pct:
	CGO_ENABLED=0 go test -cover $(PKG_DIRS)

# -------------------------------------------------------------------
# Quality targets
# -------------------------------------------------------------------
.PHONY: vet
vet:
	CGO_ENABLED=0 go vet ./...

.PHONY: lint
lint:
	@command -v staticcheck >/dev/null 2>&1 || \
		(echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

.PHONY: fmt
fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

.PHONY: fmt-check
fmt-check:
	@diff=$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*')); \
	if [ -n "$$diff" ]; then \
		echo "The following files are not gofmt-compliant:"; \
		echo "$$diff"; \
		exit 1; \
	fi
	@echo "All files are gofmt-compliant."

# -------------------------------------------------------------------
# Smoke / integration tests
# -------------------------------------------------------------------
.PHONY: smoke
smoke: build
	@echo ""
	@echo "--- true / false ---"
	./$(BINARY) true;  echo "true  exit=$$?"
	./$(BINARY) false; echo "false exit=$$?"
	@echo ""
	@echo "--- echo ---"
	./$(BINARY) echo hello world
	./$(BINARY) echo --json hello world
	@echo ""
	@echo "--- uname ---"
	./$(BINARY) uname
	./$(BINARY) uname --json
	@echo ""
	@echo "--- whoami ---"
	./$(BINARY) whoami
	./$(BINARY) whoami --json
	@echo ""
	@echo "--- pwd ---"
	./$(BINARY) pwd
	./$(BINARY) pwd --json
	@echo ""
	@echo "--- hostname ---"
	./$(BINARY) hostname
	./$(BINARY) hostname --json
	@echo ""
	@echo "--- printenv HOME ---"
	./$(BINARY) printenv HOME
	./$(BINARY) printenv --json HOME
	@echo ""
	@echo "--- env -i FOO=bar ---"
	./$(BINARY) env -i FOO=bar
	./$(BINARY) env --json -i FOO=bar
	@echo ""
	@echo "--- unknown command (expect 127) ---"
	./$(BINARY) nonexist; echo "exit=$$?"
	@echo ""
	@echo "--- --help ---"
	./$(BINARY) --help
	@echo ""
	@echo "--- --version ---"
	./$(BINARY) --version
	@echo ""
	@echo "smoke: all checks done"

.PHONY: symlink-test
symlink-test: build
	@echo "Creating symlink: echo -> korego"
	ln -sf ./$(BINARY) ./echo
	@echo "Running ./echo via symlink..."
	./echo symlink dispatch works
	./echo --json symlink json
	rm -f ./echo
	@echo "symlink-test: PASS"

# -------------------------------------------------------------------
# Housekeeping
# -------------------------------------------------------------------
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY)-race coverage.out coverage.html
	@echo "clean: done"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: all
all: vet test build
	@echo "all: vet + test + build complete"
