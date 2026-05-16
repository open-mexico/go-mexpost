---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "go-mexpost"
  text: "Microservicio de códigos postales y geocodificación de México"
  tagline: Consulta colonias, municipios y coordenadas con una API ultrarrápida construida en Go.
  actions:
    - theme: brand
      text: Inicio rápido
      link: /configuracion
    - theme: alt
      text: Referencia de endpoints
      link: /endpoints

features:
  - title: Arquitectura Hexagonal
    details: Dominio puro desacoplado de la base de datos y el framework HTTP. Fácil de testear y extender.
  - title: 145,420 colonias de México
    details: Base de datos SQLite embebida con todos los códigos postales del país, índices optimizados y polígonos GeoJSON.
  - title: Geocodificación inversa
    details: Dado un par de coordenadas GPS, devuelve la colonia postal correspondiente usando Ray-Casting sobre polígonos reales.
  - title: Sin dependencias externas
    details: Driver SQLite 100% en Go (sin CGO). Un solo binario que corre en Linux, macOS y Windows.
  - title: Rate limiting por IP
    details: Middleware token bucket integrado. Protege el servicio de abuso sin configuración adicional.
  - title: Paginación y filtros combinables
    details: Filtra por CP, nombre, municipio, geometría y combínalos libremente. Navega resultados con paginación simple.
---

## Inicio rápido

### ⚡ Opción rápida: sin compilar (Releases)

Si no deseas instalar Go, descarga el binario precompilado desde la página de [Releases](https://github.com/open-mexico/go-mexpost/releases). Elige el `.zip` de tu sistema operativo (`windows`, `mac` o `linux`), descomprímelo junto al archivo `mapa.db` e inícialo:

```bash
# Linux / macOS
chmod +x go-mexpost-api
./go-mexpost-api
```

En Windows, haz doble clic en `go-mexpost-api.exe`.

---

### 🛠️ Desde el código fuente

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

## Convención de commits

Usamos preferentemente **[Gitmoji](https://gitmoji.dev)** para los mensajes de commit. Agrega el emoji que corresponda al tipo de cambio:

| Emoji | Tipo                  |
| ----- | --------------------- |
| ✨    | Nueva función         |
| 🐛    | Corrección de bug     |
| 📝    | Documentación         |
| ♻️    | Refactor              |
| ⚡️    | Mejora de rendimiento |
| 🔧    | Configuración         |
| ✅    | Tests                 |

```bash
git commit -m '✨ Añadir endpoint de municipios'
```

## IA y Automatización

Si usarás agentes LLM para contribuir código en este repositorio:

- Sigue la guía de contribución para IA: [Guía de Agentes LLM](/agentes-llm).
- Usa el contrato de consumo API para asistentes integradores: [LLM Endpoints](/llm-endpoints).
- Ejecuta siempre el gate de calidad antes de finalizar cambios:

```bash
make verify
```

## Ejemplos rápidos

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
