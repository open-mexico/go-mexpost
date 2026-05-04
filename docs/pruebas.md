# Guía de Pruebas Unitarias

Go-MexPost sigue una estrategia **TDD (Test-Driven Development)** con pruebas en tres capas: servicios, algoritmo geoespacial y handlers HTTP. Todas las pruebas usan exclusivamente mocks (sin tocar la base de datos real), lo que las hace deterministas, rápidas y desacopladas.

---

## Estructura de los paquetes de prueba

```
internal/
├── core/services/
│   ├── colonia_srv_test.go   ← Pruebas de lógica de negocio (con MockRepo)
│   └── geometry_test.go      ← Pruebas del algoritmo Ray-Casting
└── adapters/handler/
    └── http_handler_test.go  ← Pruebas de endpoints HTTP (con MockService)
```

---

## Ejecutar las pruebas

```bash
# Todos los paquetes
go test ./...

# Un paquete específico con salida verbosa
go test -v ./internal/core/services/...
go test -v ./internal/adapters/handler/...

# Con reporte de cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Capa 1 — Servicios (`colonia_srv_test.go`)

Las pruebas de servicio verifican todas las **reglas de negocio** y los flujos de éxito/error sin depender de SQLite. Se usa un `MockRepo` que implementa la interfaz `ports.ColoniaRepository`.

### MockRepo

```go
type MockRepo struct {
    Err        error                    // error a simular
    Colonias   []domain.Colonia         // datos a devolver
    Municipios []domain.Municipio

    LastColoniaFilter ports.ColoniaSearchFilter   // captura el último filtro recibido
    LastGeoFilter     ports.ReverseGeocodeFilter
}
```

El campo `Last*Filter` permite verificar que el servicio construyó y propagó correctamente el filtro hacia el repositorio.

### Pruebas incluidas

| Test | Verifica |
|---|---|
| `TestBuscarColonias_ErroresDeValidacion` | `ValidationError` cuando falta `cp` y `nombre` |
| `TestBuscarColonias_ExitoSinGeo` | Resultado correcto y que `Geometria=nil` cuando `incluirGeo=false` |
| `TestBuscarColonias_RepoError` | Propagación de errores del repositorio |
| `TestBuscarMunicipios_NotFound` | `ErrNotFound` cuando el repo devuelve lista vacía |
| `TestBuscarPorCoordenadas_ValidaRangos` | `ValidationError` cuando `lat > 90` |
| `TestBuscarPorCoordenadas_PointInPolygon` | Punto dentro del polígono → colonia encontrada |
| `TestBuscarPorCoordenadas_NotFound` | Punto fuera del polígono → `ErrNotFound` |

### Ejemplo detallado: prueba de omisión de geometría

```go
func TestBuscarColonias_ExitoSinGeo(t *testing.T) {
    geo := `{"type":"Polygon","coordinates":[...]}`
    mockDatos := []domain.Colonia{{Codigo: "06700", Nombre: "Roma Norte", Geometria: &geo}}
    repo := &MockRepo{Colonias: mockDatos}
    servicio := services.NewColoniaService(repo)

    // Pedimos sin geometría (incluirGeo=false)
    resultado, err := servicio.BuscarColonias(ports.ColoniaSearchFilter{CP: "067"}, false)

    assert.NoError(t, err)
    assert.Len(t, resultado, 1)
    assert.Equal(t, "Roma Norte", resultado[0].Nombre)
    assert.Nil(t, resultado[0].Geometria)                   // ← geometría eliminada por el servicio
    assert.Equal(t, "067", repo.LastColoniaFilter.CP)       // ← filtro propagado correctamente
}
```

---

## Capa 2 — Algoritmo Geoespacial (`geometry_test.go`)

Prueba directamente `PointInGeoJSON` con polígonos simulados para validar que el Ray-Casting funciona correctamente en los tres casos clave.

### Polígono de prueba

Se usa un cuadrado unitario simple como geometría de referencia:

```json
{
  "type": "Polygon",
  "coordinates": [[[0,0],[10,0],[10,10],[0,10],[0,0]]]
}
```

```
    (0,10)─────(10,10)
       │           │
       │    ●(5,5) │   ← DENTRO
       │           │
    (0,0)──────(10,0)

●(15,5)               ← FUERA

●(0,5)                ← BORDE (se trata como DENTRO)
```

### Pruebas incluidas

| Test | Punto | Resultado esperado |
|---|---|---|
| `TestPointInGeoJSON_InsideSquare` | `(5, 5)` — centro del cuadrado | `true` |
| `TestPointInGeoJSON_OutsideSquare` | `(15, 5)` — a la derecha | `false` |
| `TestPointInGeoJSON_BoundarySquare` | `(0, 5)` — sobre el borde izquierdo | `true` |

### Ejemplo

```go
func TestPointInGeoJSON_InsideSquare(t *testing.T) {
    geo := `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}`

    inside, err := services.PointInGeoJSON(5, 5, geo)

    assert.NoError(t, err)
    assert.True(t, inside)
}
```

> **Nota**: `PointInGeoJSON(lon, lat, ...)`. El primer argumento es **longitud** (X), el segundo **latitud** (Y), siguiendo la convención GeoJSON.

---

## Capa 3 — Handlers HTTP (`http_handler_test.go`)

Prueba los endpoints Gin usando `net/http/httptest` para simular peticiones reales sin levantar un servidor. Se usa un `MockService` que implementa `ports.ColoniaService`.

### MockService

```go
type MockService struct {
    Err       error
    Colonias  []domain.Colonia
    Municipio []domain.Municipio
}
```

### Patrón de prueba

```go
gin.SetMode(gin.TestMode)               // 1. Modo de prueba (sin logging)
router := gin.New()                     // 2. Router limpio (sin middlewares)
router.GET("/colonias", manejador.BuscarColonias)

req, _ := http.NewRequest("GET", "/colonias?cp=067", nil)
w := httptest.NewRecorder()             // 3. Captura la respuesta en memoria
router.ServeHTTP(w, req)               // 4. Ejecuta la petición

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), `"codigo":"06700"`)
```

### Pruebas incluidas

| Test | Endpoint | Escenario | Código esperado |
|---|---|---|---|
| `TestBuscarColonias_Status400` | `GET /colonias` | `MockService` devuelve `ValidationError` | 400 |
| `TestBuscarColonias_Status200SinGeo` | `GET /colonias?cp=067` | Éxito sin `incluir_geo` | 200, sin campo `geometria` |
| `TestBuscarColonias_Status500` | `GET /colonias?cp=067` | `MockService` devuelve error genérico | 500 |
| `TestBuscarMunicipios_Status404` | `GET /municipios?estado_id=09` | `MockService` devuelve `ErrNotFound` | 404 |
| `TestBuscarCoordenadas_Status400PorParametros` | `GET /coordenadas?lat=19.4` | Falta el parámetro `lon` | 400 |
| `TestBuscarCoordenadas_Status200ConGeo` | `GET /coordenadas?lat=19.4&lon=-99.1&incluir_geo=true` | Éxito con geometría | 200, con campos `geometria` y `centro_lon` |

### Verificación de omisión de geometría

```go
func TestBuscarColonias_Status200SinGeo(t *testing.T) {
    // El mock tiene un campo geometria en el dominio...
    geo := `{"type":"Polygon",...}`
    mockService := &MockService{Colonias: []domain.Colonia{{Geometria: &geo}}}

    // ...pero el handler llama al servicio con incluirGeo=false
    req, _ := http.NewRequest("GET", "/colonias?cp=067", nil)  // sin &incluir_geo=true
    ...

    assert.Equal(t, http.StatusOK, w.Code)
    assert.NotContains(t, w.Body.String(), "geometria")  // ← no debe aparecer en JSON
}
```

---

## Estrategia de mocks

El proyecto usa mocks manuales (sin generadores automáticos) para mantener las dependencias mínimas. Cada capa mockea **solo la capa inmediatamente inferior**:

```
Handler tests    →  mockean  →  ColoniaService (interfaz)
Service tests    →  mockean  →  ColoniaRepository (interfaz)
```

Esto garantiza que cada capa se prueba de forma aislada. Si el repositorio SQLite cambia internamente, los tests de servicio y handler no se rompen.

---

## Cobertura por módulo

| Paquete | Tests existentes | Escenarios cubiertos |
|---|---|---|
| `internal/core/services` | 7 tests | Validación, éxito, no encontrado, error repo, Ray-Casting |
| `internal/core/services` (geo) | 3 tests | Dentro, fuera, borde |
| `internal/adapters/handler` | 6 tests | 200, 400, 404, 500 para los 3 endpoints |
