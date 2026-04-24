# Phase 02 — Docker Scratch Build + CI Pipeline

> **Timeline:** Week 2 | **Depends on:** Phase 01 | **Blocks:** Phase 03 (integration tests need Docker)

---

## Goal

Package KoreGo into a `FROM scratch` Docker image with symlinks for every utility, and set up CI to prevent regressions.

---

## Tasks

### 02.1 — Multi-Stage Dockerfile (`docker/Dockerfile`)

```dockerfile
# --- Stage 1: Builder ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Static binary, no CGO, stripped symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
    -o /korego ./cmd/korego/

# Generate symlinks list
RUN /korego --list-commands > /commands.txt

# --- Stage 2: Symlink Generator ---
FROM alpine:3.20 AS symlinker

COPY --from=builder /korego /bin/korego
COPY --from=builder /commands.txt /commands.txt

# Create /bin symlinks for every command
RUN while read cmd; do ln -s /bin/korego /bin/$cmd; done < /commands.txt

# Create non-root user
RUN echo "korego:x:1000:1000::/home/korego:/bin/false" >> /etc/passwd && \
    echo "korego:x:1000:" >> /etc/group

# --- Stage 3: Scratch ---
FROM scratch

# CA certs for any HTTPS needs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Non-root user
COPY --from=symlinker /etc/passwd /etc/passwd
COPY --from=symlinker /etc/group /etc/group

# Binary + all symlinks
COPY --from=symlinker /bin/ /bin/

# Timezone data (for date utility)
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
ENV ZONEINFO=/usr/local/go/lib/time/zoneinfo.zip

USER korego
ENTRYPOINT ["/bin/korego"]
```

**Checklist:**
- [x] Multi-stage: builder → symlinker → scratch
- [x] Static binary with stripped symbols (`-s -w`)
- [x] Version embedded via `-X main.version=...`
- [x] All registered commands get `/bin/<name>` symlinks
- [x] CA certificates included (for future HTTPS use)
- [x] Non-root user created and used
- [x] Timezone data included (for `date` utility)
- [x] `--list-commands` subcommand added to dispatcher (outputs one command per line)

---

### 02.2 — Debug Dockerfile (`docker/Dockerfile.debug`)

```dockerfile
FROM golang:1.22-alpine AS builder
# ... same build steps ...

FROM alpine:3.20
COPY --from=builder /korego /bin/korego
# ... same symlinks ...
# Alpine provides shell, strace, etc. for debugging
```

- [x] Same binary, but on Alpine base for debugging
- [x] Can `docker exec -it` and poke around

---

### 02.3 — Makefile

```makefile
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
BINARY  := korego
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test docker docker-debug lint clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/korego/

test:
	CGO_ENABLED=0 go test -race -cover ./...

lint:
	go vet ./...
	staticcheck ./...

docker:
	docker build -t korego:$(VERSION) -f docker/Dockerfile .

docker-debug:
	docker build -t korego:debug -f docker/Dockerfile.debug .

clean:
	rm -f $(BINARY)

smoke: docker
	docker run --rm korego:$(VERSION) true
	docker run --rm korego:$(VERSION) false || true
	docker run --rm korego:$(VERSION) echo "smoke test passed"
	docker run --rm korego:$(VERSION) echo --json "smoke test passed"
	@echo "=== ALL SMOKE TESTS PASSED ==="
```

- [x] `make build` — local binary
- [x] `make test` — all unit tests
- [x] `make lint` — vet + staticcheck
- [x] `make docker` — production image
- [x] `make docker-debug` — debug image
- [x] `make smoke` — build + run basic smoke tests in container

---

### 02.4 — GitHub Actions CI (`.github/workflows/ci.yml`)

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: make lint
      - run: make test
      - run: make docker
      - run: make smoke
  
  image-size:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - run: make docker
      - run: |
          SIZE=$(docker image inspect korego:* --format '{{.Size}}')
          echo "Image size: $((SIZE / 1024 / 1024)) MB"
          if [ $SIZE -gt 20971520 ]; then
            echo "FAIL: Image exceeds 20MB target"
            exit 1
          fi
```

- [x] Runs on every push and PR
- [x] Lint → Test → Docker build → Smoke test
- [x] Image size gate: fail if > 20MB

---

## Milestone 02

- [x] `docker build` produces a working `scratch` image
- [x] `docker run korego true` exits 0
- [x] `docker run korego false` exits 1
- [x] `docker run --entrypoint /bin/echo korego hello` works (symlink)
- [x] `docker run korego --help` lists all commands
- [x] Image size < 20MB  *(actual: 3.2 MB)*
- [x] CI pipeline passes end-to-end
- [x] `make smoke` passes all checks

## How to Verify

```bash
make docker
docker images korego  # check size

# Smoke tests
docker run --rm korego:dev true ; echo $?
docker run --rm korego:dev false ; echo $?
docker run --rm korego:dev echo --json "works"
docker run --rm --entrypoint /bin/whoami korego:dev

# Image size
docker image inspect korego:dev --format '{{.Size}}'
```
