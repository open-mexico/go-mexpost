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

## Estructura de `cmd/api/main.go`

El punto de entrada del servidor realiza el **Wiring** (inyección de dependencias manual) en orden:

```go
// 1. Repositorio — conecta con mapa.db
repo, err := repository.NewSQLiteRepository("./mapa.db")

// 2. Servicio — recibe el repositorio como interfaz
servicio := services.NewColoniaService(repo)

// 3. Handler — recibe el servicio como interfaz
apiHandler := handler.NewHttpHandler(servicio)

// 4. Router Gin — registra los endpoints
router := gin.Default()
router.GET("/colonias",    apiHandler.BuscarColonias)
router.GET("/municipios",  apiHandler.BuscarMunicipios)
router.GET("/coordenadas", apiHandler.BuscarCoordenadas)

router.Run(":8080")
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

---

## Variables de entorno

Actualmente el servidor no requiere configuración por variables de entorno. La ruta de la base de datos está hardcodeada como `./mapa.db` (relativa al directorio de trabajo).

Para personalizar la ruta en un futuro, el punto de extensión natural sería leer la ruta desde `os.Getenv("DB_PATH")` en `cmd/api/main.go` antes de llamar a `NewSQLiteRepository`.

---

## Árbol de dependencias clave (`go.mod`)

```
github.com/open-mexico/go-mexpost
├── github.com/gin-gonic/gin v1.12.0       ← HTTP framework
├── modernc.org/sqlite v1.50.0             ← SQLite sin CGO
└── github.com/stretchr/testify v1.11.1   ← Testing
```

---

## Verificación de instalación

Una vez iniciado el servidor, puedes verificar que todo funciona con estas peticiones de prueba:

```bash
# Buscar colonias por CP
curl "http://localhost:8080/colonias?cp=06700"

# Buscar municipios por estado
curl "http://localhost:8080/municipios?estado_id=09"

# Geocodificación inversa (Roma Norte, CDMX)
curl "http://localhost:8080/coordenadas?lat=19.4181&lon=-99.1634"

# Con polígono GeoJSON
curl "http://localhost:8080/coordenadas?lat=19.4181&lon=-99.1634&incluir_geo=true"
```
