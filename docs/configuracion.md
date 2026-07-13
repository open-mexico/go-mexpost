# Configuración e Instalación

## Requisitos previos

| Herramienta | Versión mínima | Instalación |
|---|---|---|
| Go | 1.26 | [go.dev/dl](https://go.dev/dl/) |
| Git | cualquiera | [git-scm.com](https://git-scm.com/) |
| Conexión a internet | — | Para descargar la base de datos |

No se requiere instalar ningún servidor de base de datos. SQLite está embebido en el binario.

---

## Instalación paso a paso

### 1. Clonar el repositorio

```bash
git clone https://github.com/open-mexico/go-mexpost.git
cd go-mexpost
```

### 2. Instalar dependencias de Go

```bash
go mod tidy
```

Esto descargará todos los módulos declarados en `go.mod`:

- `github.com/gin-gonic/gin` — Framework HTTP
- `modernc.org/sqlite` — Driver SQLite (sin CGO)
- `github.com/stretchr/testify` — Assertions y mocks para tests

### 3. Descargar la base de datos

El proyecto incluye un comando `setup` que descarga y descomprime `mapa.db` automáticamente desde el repositorio ETL oficial.

#### Opción A — Solo datos postales (más ligera, ~12 MB)

Incluye todos los campos de texto pero **sin polígonos GeoJSON**. Los endpoints `/colonias` y `/municipios` funcionan al 100%. El endpoint `/coordenadas` devuelve 404 siempre (no hay geometrías).

```bash
go run ./cmd/setup/main.go -geo=false
```

#### Opción B — Datos + geometrías GeoJSON (recomendada, ~180 MB)

Incluye los polígonos GeoJSON de cada colonia. Habilita la geocodificación inversa en `/coordenadas`.

```bash
go run ./cmd/setup/main.go
# El flag -geo es true por defecto
```

El script realiza tres operaciones:
1. Descarga el archivo `.zip` desde GitHub Releases a `temp.zip`.
2. Extrae el archivo SQLite del ZIP.
3. Lo renombra a `mapa.db` en la raíz del proyecto.

### 4. Iniciar el servidor API

```bash
go run ./cmd/api/main.go
```

Salida esperada:
```
[GIN-debug] [WARNING] Creating an engine instance with the Logger and Recovery middleware already attached.
[GIN-debug] GET    /colonias                 --> ...
[GIN-debug] GET    /municipios               --> ...
[GIN-debug] GET    /coordenadas              --> ...
🚀 Servidor corriendo en http://localhost:8080
```

La API estará disponible en `http://localhost:8080`.

---

## Compilar el binario de producción

```bash
# Compilar para la plataforma actual
go build -o go-mexpost ./cmd/api/main.go

# Compilar para Linux (desde cualquier OS)
GOOS=linux GOARCH=amd64 go build -o go-mexpost-linux ./cmd/api/main.go

# Ejecutar
./go-mexpost
```

Al ser CGO-free (gracias a `modernc.org/sqlite`), el binario resultante no tiene dependencias dinámicas y se puede desplegar en cualquier servidor Linux sin instalar nada adicional.

---

## Estructura de `cmd/api/main.go` (Production-Ready)

El punto de entrada del servidor está optimizado para producción con apagado elegante (Graceful Shutdown) y middleware avanzado. El **Wiring** (inyección de dependencias) ocurre en el siguiente orden:

```go
// 1. Repositorio — conecta con mapa.db
repo, err := repository.NewSQLiteRepository("./mapa.db")

// 2. Servicio — recibe el repositorio como interfaz
servicio := services.NewColoniaService(repo)

// 3. Handler — recibe el servicio como interfaz
apiHandler := handler.NewHttpHandler(servicio)

// 4. Router Gin — en modo "New" (sin middlewares globales por defecto)
router := gin.New()
router.SetTrustedProxies(nil)

// 5. Middlewares de Producción
router.Use(gin.Recovery())                                  // Previene caídas por panics
router.Use(handler.SlogMiddleware(logger))                  // Logs estructurados en JSON
router.Use(cors.New(corsConfig))                            // CORS (Controlado por CORS_ALLOWED_ORIGINS)
router.Use(gzip.Gzip(gzip.DefaultCompression))              // Compresión GZIP automática
router.Use(handler.RateLimitMiddleware(rateLimit, burst))   // Token bucket por IP

// 6. Endpoints
router.GET("/health",      apiHandler.HealthCheck)
router.GET("/colonias",    apiHandler.BuscarColonias)
router.GET("/municipios",  apiHandler.BuscarMunicipios)
router.GET("/coordenadas", apiHandler.BuscarCoordenadas)

// 7. Servidor HTTP Resiliente (Timeouts para prevenir ataques Slowloris)
srv := &http.Server{
    Addr:         ":" + port,
    Handler:      router,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

// 8. Graceful Shutdown (Intercepta SIGINT y SIGTERM)
// ...
```

### Configuración CORS

Por defecto, la API permite peticiones desde cualquier origen (`*`) para que desarrolladores frontend puedan consumirla sin bloqueos de navegador. 
En **Producción**, deberías restringir esto para proteger tu infraestructura usando la variable de entorno `CORS_ALLOWED_ORIGINS`:

```bash
# Ejemplo estricto para producción
CORS_ALLOWED_ORIGINS=https://mi-app.com,https://dashboard.mi-app.com go run ./cmd/api/main.go
```

Las dependencias se inyectan como **interfaces**, no como implementaciones concretas. Esto significa que se puede reemplazar SQLite por otra base de datos implementando `ports.ColoniaRepository` sin tocar el servicio ni los handlers.

---

## Comandos útiles durante el desarrollo

```bash
# Ejecutar todos los tests
go test ./...

# Tests con salida verbosa
go test -v ./...

# Tests con cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out   # abre en el navegador

# Formatear el código
gofmt -w ./cmd ./internal

# Verificar errores de compilación sin ejecutar
go build ./...

# Ver dependencias del módulo
go mod graph

# Lint (requiere golangci-lint instalado)
golangci-lint run
```
