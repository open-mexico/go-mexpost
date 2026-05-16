# Guia de Agentes LLM para Contribucion

Esta guia define como deben trabajar los agentes de IA que hacen cambios de codigo en go-mexpost.

> Si buscas un contrato para consumir la API desde un asistente, revisa la guia de integracion en `llm-endpoints`.

## Objetivo

Mantener cambios seguros, pequenos y verificables, respetando contratos de endpoints y arquitectura hexagonal.

## Flujo minimo obligatorio

Antes de dar por finalizado cualquier cambio, ejecutar:

```bash
make fmt
make build
make test
make lint
make verify
```

Si no existe `golangci-lint` en el entorno:

```bash
brew install golangci-lint
# o
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Reglas de implementacion

1. No romper contratos publicos de los endpoints existentes.
2. Mantener la logica de negocio en servicios (`internal/core/services`).
3. Mantener handlers HTTP delgados (`internal/adapters/handler`).
4. Propagar cambios de contrato a documentacion:
   - `docs/endpoints.md`
   - `README.md` (si cambia instalacion, uso o contribucion)
5. No agregar dependencias sin justificar valor tecnico claro.

## Definicion de terminado

Un cambio esta listo cuando:

- Compila correctamente.
- Todos los tests pasan.
- Lint pasa.
- La documentacion quedo actualizada.
- Se incluyeron/ajustaron pruebas cuando hubo cambio de comportamiento.

## Mapa rapido del proyecto para agentes

- API server: `cmd/api/main.go`
- Setup de DB: `cmd/setup/main.go`
- Dominio: `internal/core/domain`
- Puertos: `internal/core/ports`
- Servicios: `internal/core/services`
- Adaptadores HTTP: `internal/adapters/handler`
- Repositorio SQLite: `internal/adapters/repository`

## Buenas practicas al editar

- Cambios pequenos y orientados a una responsabilidad.
- Validaciones de entrada explicitas.
- Mensajes de error claros.
- Evitar refactors amplios no solicitados.
- Actualizar ejemplos cuando el comportamiento observable cambia.

## Enlaces relacionados

- [Referencia de Endpoints](/endpoints)
- [Guia de Pruebas](/pruebas)
- [Guia de Integracion LLM para consumo de API](/llm-endpoints)
