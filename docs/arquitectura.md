# Arquitectura Hexagonal de go-mexpost

## ¿Qué es la Arquitectura Hexagonal?

La Arquitectura Hexagonal (también llamada _Ports & Adapters_) organiza el código en capas concéntricas con una regla fundamental: **las dependencias siempre apuntan hacia adentro**. El núcleo de la aplicación (dominio y servicios) no sabe nada de bases de datos, HTTP ni frameworks externos.

```
  ┌─────────────────────────────────────────────────────────────┐
  │                        ADAPTADORES                          │
  │                                                             │
  │   ┌──────────────┐              ┌───────────────────────┐   │
  │   │  HTTP Handler│              │   SQLite Repository   │   │
  │   │  (Gin-Gonic) │              │   (modernc.org/sqlite)│   │
  │   └──────┬───────┘              └──────────┬────────────┘   │
  │          │                                 │                │
  │   ────── │ ───── PUERTOS (interfaces) ─── │ ──────         │
  │          │                                 │                │
  │   ┌──────▼─────────────────────────────────▼────────────┐  │
  │   │                   SERVICIOS (casos de uso)           │  │
  │   │   BuscarColonias · BuscarMunicipios · BuscarCoords   │  │
  │   └────────────────────────┬────────────────────────────┘  │
  │                            │                                │
  │   ┌────────────────────────▼────────────────────────────┐  │
  │   │                     DOMINIO                          │  │
  │   │   Colonia · Municipio · ErrNotFound · ValidationErr  │  │
  │   └─────────────────────────────────────────────────────┘  │
  └─────────────────────────────────────────────────────────────┘
```

---

## Estructura de directorios

```
go-mexpost/
├── cmd/
│   ├── api/
│   │   └── main.go          ← Punto de entrada del servidor HTTP
│   └── setup/
│       └── main.go          ← Descarga y prepara la base de datos
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   └── colonia.go   ← Modelos y errores de negocio
│   │   ├── ports/
│   │   │   └── ports.go     ← Interfaces (contratos) de repositorio y servicio
│   │   └── services/
│   │       ├── colonia_srv.go     ← Lógica de negocio de los tres casos de uso
│   │       ├── geometry.go        ← Algoritmo Ray-Casting (Point-in-Polygon)
│   │       ├── colonia_srv_test.go
│   │       └── geometry_test.go
│   └── adapters/
│       ├── handler/
│       │   ├── http_handler.go    ← Endpoints Gin: /colonias /municipios /coordenadas
│       │   ├── ratelimit.go       ← Middleware de rate limiting por IP (token bucket)
│       │   └── http_handler_test.go
│       └── repository/
│           └── sqlite_repo.go     ← Queries SQL dinámicas contra mapa.db
├── docs/                    ← Esta documentación
├── go.mod
├── go.sum
└── mapa.db                  ← Base de datos SQLite (generada por cmd/setup)
```

---

## Capas explicadas

### 1. Dominio (`internal/core/domain/`)

Es el núcleo puro del sistema. Contiene estructuras de datos y errores tipados que representan conceptos del negocio. **No importa ningún paquete externo** salvo la librería estándar de Go.

| Tipo | Descripción |
|---|---|
| `Colonia` | Código postal, nombre, tipo, ciudad, zona, estado, municipio, BBox y centroide |
| `Municipio` | ID, nombre y estado al que pertenece |
| `ErrNotFound` | Error centinela: búsqueda sin resultados → HTTP 404 |
| `ValidationError` | Error de reglas de negocio → HTTP 400 |

Los campos de `Colonia` como `Geometria *string`, `CentroLon *float64` y `CentroLat *float64` son punteros porque pueden ser `NULL` en la base de datos.

---

### 2. Puertos (`internal/core/ports/`)

Define los **contratos** (interfaces de Go) que separan el núcleo de las implementaciones concretas. Esto permite reemplazar SQLite por PostgreSQL, o Gin por net/http, sin tocar una línea de lógica de negocio.

#### Filtros de búsqueda

```go
// Para GET /colonias
type ColoniaSearchFilter struct {
    CP          string  // Código postal (prefijo o exacto, 3-5 dígitos)
    Nombre      string  // Nombre de colonia (parcial o exacto)
    MunicipioID string  // Filtro adicional por municipio
    SoloGeo     bool    // true → solo colonias con geometría
    Limit       int     // Máximo de resultados (default 100, máx 500)
    Offset      int     // Desplazamiento para paginación
}

// Para GET /municipios
type MunicipioSearchFilter struct {
    Nombre   string  // Nombre del municipio
    EstadoID string  // Filtro por estado
}

// Para GET /coordenadas
type ReverseGeocodeFilter struct {
    Lat      float64 // Latitud del punto
    Lon      float64 // Longitud del punto
    EstadoID string  // Filtro opcional por estado
}
```

#### ColoniaRepository (adaptador de salida)

```go
type ColoniaRepository interface {
    SearchColonias(filter ColoniaSearchFilter) ([]domain.Colonia, error)
    CountColonias(filter ColoniaSearchFilter) (int, error)
    SearchMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
    FindColoniasByPointBBox(filter ReverseGeocodeFilter) ([]domain.Colonia, error)
}
```

#### ColoniaService (adaptador de entrada)

```go
type ColoniaService interface {
    BuscarColonias(filter ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error)
    ContarColonias(filter ColoniaSearchFilter) (int, error)
    BuscarMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
    BuscarPorCoordenadas(filter ReverseGeocodeFilter, incluirGeo bool) (*domain.Colonia, error)
}
```

---

### 3. Servicios (`internal/core/services/`)

Implementan los **casos de uso** del sistema. Contienen toda la lógica de negocio, validaciones y orquestación. Los servicios solo hablan con el repositorio a través de la interfaz `ColoniaRepository`.

#### `BuscarColonias`
1. Sanitiza los strings de entrada (trim de espacios).
2. Valida que se proporcione al menos `cp` o `nombre` → `ValidationError` si no.
3. Valida que `cp` tenga entre 3 y 5 dígitos y que sean solo dígitos → `ValidationError` si no cumple.
4. Aplica límites de `Limit`/`Offset` según el modo (`incluirGeo` o no): default 100/50, máx 500/100.
5. Llama a `repo.SearchColonias(filter)` con `LIMIT` y `OFFSET`.
6. Si no hay resultados → `ErrNotFound`.
7. Si `incluirGeo = false`, pone a `nil` los campos `Geometria`, `CentroLon`, `CentroLat` antes de devolver.

#### `ContarColonias`
1. Sanitiza y valida el filtro igual que `BuscarColonias`.
2. Llama a `repo.CountColonias(filter)` (sin LIMIT/OFFSET) para obtener el total real.
3. El handler usa este total para calcular `total_paginas` y construir las URLs de navegación.

#### `BuscarMunicipios`
1. Sanitiza `nombre` y `estado_id`.
2. Valida que se proporcione al menos uno → `ValidationError` si no.
3. Llama a `repo.SearchMunicipios(filter)`.
4. Si no hay resultados → `ErrNotFound`.

#### `BuscarPorCoordenadas`
1. Valida rangos: `lat ∈ [-90, 90]`, `lon ∈ [-180, 180]`.
2. Llama a `repo.FindColoniasByPointBBox(filter)` (prefiltrado rápido por BBox).
3. Para cada candidata con geometría, ejecuta el algoritmo **Ray-Casting** (`PointInGeoJSON`).
4. Devuelve la primera colonia cuyo polígono contiene el punto.
5. Si ninguna lo contiene → `ErrNotFound`.

---

### 4. Repositorio (`internal/adapters/repository/`)

Implementa `ColoniaRepository` usando SQLite. Construye queries dinámicas de forma segura con parámetros posicionales (`?`) para evitar SQL injection.

Ver [base-de-datos](./base-de-datos) para el detalle de las queries y el esquema.

---

### 5. Handler HTTP (`internal/adapters/handler/`)

Implementa los endpoints con Gin-Gonic. Sus responsabilidades son:

- Parsear y validar parámetros de query string.
- Traducir errores del servicio a códigos HTTP.
- Dar forma al JSON de respuesta (incluyendo u omitiendo geometría).
- Construir la metadata de paginación (`pagina`, `total_paginas`, `pagina_anterior`, `pagina_siguiente`).

El middleware de rate limiting (`ratelimit.go`) se registra globalmente en el router y es ajeno al handler de negocio:

```go
// Token bucket: rateLimit req/s sostenidos, burst de rateBurst por IP.
router.Use(handler.RateLimitMiddleware(rateLimit, rateBurst))
```

Ver [endpoints](./endpoints) para la referencia completa de la API.

---

## Flujo de una petición (ejemplo: GET /colonias?cp=067)

```
Cliente HTTP
     │
     ▼
[ Gin Router ]
     │ llama a
     ▼
[ HttpHandler.BuscarColonias ]
  - parsea ?cp=067, ?incluir_geo=false
  - construye ColoniaSearchFilter{CP: "067"}
     │ llama a
     ▼
[ coloniaService.BuscarColonias(filter, false) ]
  - valida que CP no esté vacío ✓
  - llama a repo.SearchColonias(filter)
     │
     ▼
[ sqliteRepo.SearchColonias(filter) ]
  - construye: SELECT ... FROM colonias WHERE codigo LIKE '067%'
  - ejecuta y escanea filas
  - devuelve []domain.Colonia
     │
     ▼ (de vuelta al servicio)
  - len > 0 ✓
  - incluirGeo=false → Geometria=nil, CentroLon=nil, CentroLat=nil
  - devuelve []domain.Colonia (sin geo)
     │
     ▼ (de vuelta al handler)
  - mapea a []coloniaResponse (struct con json tags)
  - responde 200 OK con {"resultados": [...]}
     │
     ▼
Cliente HTTP ← {"resultados": [...]}
```

---

## Mapeo de errores a HTTP

| Error en servicio | Código HTTP | Cuerpo |
|---|---|---|
| `domain.ValidationError` | `400 Bad Request` | `{"error": "parametros invalidos", "detalle": "<mensaje>"}` |
| `domain.ErrNotFound` | `404 Not Found` | `{"error": "no se encontraron resultados", "detalle": "ajusta tus filtros e intenta de nuevo"}` |
| Cualquier otro error | `500 Internal Server Error` | `{"error": "error interno", "detalle": "ocurrio un error procesando la solicitud"}` |
| Rate limit excedido | `429 Too Many Requests` | `{"error": "demasiadas solicitudes", "detalle": "excediste el límite de solicitudes, intenta de nuevo en unos segundos"}` |
