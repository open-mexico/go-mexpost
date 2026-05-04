# Documentación de go-mexpost

Microservicio ultrarrápido para consulta de colonias, códigos postales y geocodificación inversa de México, construido con Go 1.26 y Arquitectura Hexagonal.

---

## Índice de documentos

| Documento | Descripción |
|---|---|
| [arquitectura.md](arquitectura.md) | Diseño hexagonal, capas, flujo de una petición y mapeo de errores a HTTP |
| [endpoints.md](endpoints.md) | Referencia completa de los 3 endpoints con parámetros, combinaciones y ejemplos JSON |
| [geocodificacion.md](geocodificacion.md) | Explicación del algoritmo Ray-Casting y la estrategia de prefiltrado BBox |
| [base-de-datos.md](base-de-datos.md) | Esquema de tablas, índices, queries dinámicas y versiones de `mapa.db` |
| [pruebas.md](pruebas.md) | Estrategia de testing por capas, mocks, casos cubiertos y cómo ejecutar los tests |
| [configuracion.md](configuracion.md) | Instalación, compilación, comandos de desarrollo y verificación |

---

## Resumen rápido

```
GET /colonias?cp=067
GET /colonias?nombre=Roma&incluir_geo=true
GET /colonias?cp=067&nombre=Roma&municipio_id=014
GET /colonias?solo_geo=true&cp=067

GET /municipios?estado_id=14
GET /municipios?nombre=Zapopan

GET /coordenadas?lat=19.4181&lon=-99.1634
GET /coordenadas?lat=19.4181&lon=-99.1634&estado_id=09&incluir_geo=true
```

## Inicio rápido

```bash
# 1. Clonar
git clone https://github.com/open-mexico/go-mexpost.git && cd go-mexpost

# 2. Dependencias
go mod tidy

# 3. Descargar base de datos (con geometrías GeoJSON)
go run ./cmd/setup/main.go

# 4. Correr tests
go test ./...

# 5. Iniciar servidor
go run ./cmd/api/main.go
# → http://localhost:8080
```
