# AGENTS Guide - go-mexpost

This file defines the rules for LLM agents that contribute code to this repository.

## 1. Agent Objective

- Maintain API stability and compatibility with documented contracts.
- Prioritize small, testable, and easy-to-review changes.
- Update documentation when observable behavior changes.

## 2. Project Context

- Main language: Go 1.26.
- Architecture: Hexagonal (domain + ports + adapters).
- API entrypoint: cmd/api/main.go.
- Database setup entrypoint: cmd/setup/main.go.
- Main layers:
  - Domain: internal/core/domain
  - Ports: internal/core/ports
  - Services: internal/core/services
  - HTTP adapters: internal/adapters/handler
  - SQLite repository: internal/adapters/repository

## 3. Change Rules

1. Preserve public endpoint contracts unless explicitly requested otherwise.
2. Keep separation by layers; avoid business logic in handlers.
3. Do not couple services to infrastructure details.
4. If you add request parameters or response fields:
   - Update docs in the docs folder.
   - Update README if usage/installation/contribution is affected.
5. Do not introduce unnecessary dependencies if a simple stdlib alternative exists.

## 4. Mandatory Quality Flow

Run these steps before finishing any change:

```bash
# 1) Formatting
make fmt

# 2) Build validation
make build

# 3) Unit tests
make test

# 4) Static linting
make lint

# 5) Full verification
make verify
```

If your environment does not have golangci-lint:

```bash
# Recommended installation
brew install golangci-lint
# or
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## 5. Definition of Done (DoD)

A change is considered done only if:

- It compiles without errors.
- All tests pass.
- Lint passes without errors.
- Documentation is updated (docs and/or README when applicable).
- It does not break published usage examples.

## 6. Pull Request Checklist for Agents

- [ ] Change is scoped to the objective.
- [ ] No unrelated files or code included.
- [ ] New or updated tests included when behavior changes.
- [ ] Documentation references are updated.
- [ ] Risks and limitations are documented in the PR description.

## 7. Recommended Conventions

- Use clear names in Spanish or English, but keep consistency inside each module.
- Return errors with useful debugging context.
- Keep input validation at system boundaries (handlers/services as appropriate).
- Keep functions small and focused on a single responsibility.

## 8. Guardrails for LLM Agents

- Do not invent endpoints or fields outside documented contracts.
- Do not assume geometry exists if the database was downloaded without -geo.
- Avoid massive style-only changes without functional value.
- If behavior is unclear, prioritize consistency with existing tests.

## 9. References

- docs/endpoints.md
- docs/pruebas.md
- docs/configuracion.md
- docs/llm-endpoints.md

## 10. Estrategia de Dockerización

- **Un solo Dockerfile parametrizado:** Utiliza argumentos (`ARG`) para controlar las variables de compilación en lugar de crear múltiples archivos `Dockerfile` (ej. `Dockerfile.lite`, `Dockerfile.geo`).
- **Autonomía en el Build:** El contenedor debe generar su propia base de datos embebida durante la etapa de construcción (`builder stage`) usando `go run ./cmd/setup/main.go`. No asumas que la base de datos `mapa.db` existe en la máquina anfitriona.
- **Registro en el Makefile:** Todas las variantes de construcción de contenedores deben documentarse como targets explícitos en el `Makefile` (ej. `make docker-lite`, `make docker-geo`).

## 11. Contexto de Usuario
- **Usuario Oficial:** El usuario oficial para Docker Hub, GitHub y cualquier otro servicio externo es `macarthuror`. 
- **Documentación y Comandos:** Cuando se generen comandos de terminal, `Dockerfile`, scripts o documentación que requieran un nombre de usuario o espacio de nombres, utiliza siempre `macarthuror` (ej. `docker pull macarthuror/go-mexpost:lite`) en lugar de placeholders genéricos.
