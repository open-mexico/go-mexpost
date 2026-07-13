.PHONY: help fmt fmt-check build test lint verify docs-build

GO_DIRS := ./cmd ./internal

help:
	@echo "Targets disponibles:"
	@echo "  make fmt         - Formatea el codigo Go"
	@echo "  make fmt-check   - Verifica formato Go sin modificar archivos"
	@echo "  make build       - Verifica compilacion"
	@echo "  make test        - Ejecuta tests"
	@echo "  make lint        - Ejecuta golangci-lint"
	@echo "  make verify      - Ejecuta fmt-check + build + test + lint"
	@echo "  make docs-build  - Compila documentacion (VitePress)"
	@echo "  make docker-lite - Construye imagen Docker ligera (sin polígonos)"
	@echo "  make docker-geo  - Construye imagen Docker espacial (con polígonos)"

fmt:
	@gofmt -w $(GO_DIRS)

fmt-check:
	@test -z "$$(gofmt -l $(GO_DIRS))" || (echo "Hay archivos sin formatear. Ejecuta: make fmt"; gofmt -l $(GO_DIRS); exit 1)

build:
	@go build ./...

test:
	@go test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint no esta instalado. Instala con: brew install golangci-lint"; exit 1)
	@golangci-lint run

verify: fmt-check build test lint

docs-build:
	@pnpm docs:build

docker-lite:
	@echo "Construyendo Docker LITE (sin geometrías)..."
	@docker build --build-arg INCLUDE_GEO=false -t go-mexpost:lite .

docker-geo:
	@echo "Construyendo Docker GEO (con geometrías)... Esto puede tomar varios minutos."
	@docker build --build-arg INCLUDE_GEO=true -t go-mexpost:geo .
