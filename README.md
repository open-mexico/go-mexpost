[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<br />
<p align="center">
  <h1 align="center">Go-MexPost 🇲🇽🚀</h1>

  <!-- <div align="center">
    <img align="center" src="assets/mex-post.png" alt="MexPost Logo">
  </div> -->

  <p align="center">
    Microservicio ultrarrápido para consulta de colonias, códigos postales y coordenadas de México.
    <br />
    <a href="https://github.com/open-mexico/go-mexpost/issues">Report Bug</a>
    ·
    <a href="https://github.com/open-mexico/go-mexpost/issues">Request Feature</a>
  </p>
</p>

---

## 🔑 Abstract (English)

**Go-MexPost** is a blazing-fast, open-source microservice built in Golang, designed to simplify access to postal and geographic data in Mexico. Formerly built on Node.js, this project was completely rewritten from scratch using a strict **Hexagonal Architecture** and an embedded SQLite database to achieve extreme performance, high concurrency, and minimal memory footprint. 

The data generation process has been decoupled into an external ETL pipeline, making this API fully agnostic and effortlessly maintainable. Users can search by postal code, partial names, or perform reverse geocoding to find neighborhoods based on geographic coordinates using GeoJSON polygons. 

## 📍 Descripción

**Go-MexPost** es la evolución de nuestro proyecto original. Hemos reescrito el motor por completo utilizando **Golang** para ofrecer un microservicio ultrarrápido, capaz de manejar miles de peticiones por segundo consumiendo muy poca memoria RAM.

Para facilitar su mantenimiento, hemos separado la construcción de la base de datos a un [repositorio externo (ETL)](https://github.com/open-mexico/sepomex-db-generator). Esto permite que Go-MexPost funcione mediante una base de datos embebida (SQLite), lo que significa que puedes desplegar este microservicio en cualquier servidor en segundos, sin necesidad de configurar gestores de bases de datos complejos.

## 🎯 Objetivo

Proveer una API REST robusta, eficiente y de nivel empresarial para desarrolladores, investigadores y empresas que requieran consultar y gestionar datos de colonias en México de forma instantánea. Todo esto respaldado por un diseño modular, pruebas unitarias exhaustivas y código limpio.

## 🔍 Características

* ⚡ **Rendimiento Extremo:** Gracias a Go y Gin-Gonic, las respuestas se sirven en milisegundos.
* 📦 **Base de Datos Embebida (SQLite):** No requiere instalar MySQL ni PostgreSQL. Descarga la base de datos lista para usar y arranca el servidor.
* 🔎 **Búsquedas Flexibles y Optimizadas:**
  * Por Código Postal exacto o parcial (ej. `067%`).
  * Por Nombre de Colonia exacto o parcial (ignorando mayúsculas y minúsculas).
  * Filtros combinados por Estado y Municipio.
* 📄 **Paginación completa:** Todos los resultados de `/colonias` incluyen `pagina`, `total_paginas`, `pagina_anterior` y `pagina_siguiente` para navegar grandes conjuntos de datos.
* 🛡️ **Rate Limiting por IP:** Protección integrada contra abuso mediante token bucket. Configurable con variables de entorno (`RATE_LIMIT`, `RATE_BURST`) sin recompilar.
* 🗺️ **Soporte Geoespacial (GeoJSON):** Capacidad de descargar una versión espacial de la base de datos que incluye los polígonos de cada código postal.
* 📍 **Geocodificación Inversa:** Búsqueda de colonias mediante coordenadas (Point-in-Polygon) y obtención de centroides geográficos.
* 🏗️ **Arquitectura Hexagonal:** Código desacoplado, testeable y preparado para escalar.

## 🚀 Instalación y Uso

Configurar el proyecto en tu máquina local es increíblemente sencillo gracias a nuestro script de inicialización integrado.

---

### ⚡ Opción Rápida: Probar sin programar (Releases)

Si no deseas instalar Go o compilar el código por tu cuenta, hemos preparado versiones precompiladas listas para usarse. ¡No requieren configuración!

**Paso 1:** Ve a nuestra página de [Releases](https://github.com/open-mexico/go-mexpost/releases).
**Paso 2:** Descarga el archivo `.zip` correspondiente a tu sistema operativo (`windows`, `mac`, o `linux`).
**Paso 3:** Descomprime el archivo. Asegúrate de que tanto el ejecutable como el archivo de la base de datos (`mapa.db`) estén en la misma carpeta.
**Paso 4:** Inicia el servidor:

* **🪟 Windows:** Haz doble clic sobre el archivo `go-mexpost-api.exe`.
* **🐧 Linux / 🍏 Mac:** Abre una terminal en esa carpeta, otorga permisos de ejecución y arráncalo con estos comandos:
  ```bash
  chmod +x go-mexpost-api
  ./go-mexpost-api

---

### 1. Clonar el repositorio
```bash
git clone https://github.com/open-mexico/go-mexpost.git
cd go-mexpost
```

### 2. Instalar dependencias

```bash
go mod tidy
```

### 3. Descargar la Base de Datos
Go-MexPost cuenta con un comando especial para descargar la base de datos oficial (generada en nuestro repo ETL) directamente a tu proyecto.

Opción A (Ligera): Solo datos postales y de texto (ideal para formularios).

```bash
go run ./cmd/setup/main.go
```

Opción B (Espacial) - **Recomendada**: Datos postales + Polígonos GeoJSON (ideal para mapas).

```bash
go run ./cmd/setup/main.go -geo=true
```

### 4. Iniciar el Servidor
```bash
go run ./cmd/api/main.go
```

La API estará corriendo en http://localhost:8080.

### 5. Variables de entorno (opcionales)

| Variable | Default | Descripción |
|---|---|---|
| `PORT` | `8080` | Puerto en el que escucha el servidor |
| `RATE_LIMIT` | `10` | Solicitudes por segundo permitidas por IP |
| `RATE_BURST` | `30` | Ráfaga máxima instantánea por IP |

```bash
# Ejemplo con configuración personalizada
PORT=3000 RATE_LIMIT=5 RATE_BURST=15 go run ./cmd/api/main.go
```

🛠️ Herramientas y Tecnologías
`Golang` (1.26+): Lenguaje principal.

Gin-Gonic: Framework web HTTP de alto rendimiento.

SQLite (modernc.org): Driver de base de datos embebida 100% en Go (CGO-free).

Testify: Herramientas para pruebas unitarias (Mocks y Asserts).

golang.org/x/time/rate: Token bucket para rate limiting por IP.

## 📊 Fuente de Datos
La información es extraída, limpiada y optimizada desde las fuentes oficiales del Servicio Postal Mexicano (SEPOMEX/Correos de México) y cruzada con límites espaciales de Open Mexico GeoJSON.

## 🔑 Palabras Clave
- Golang
- Microservicio
- Códigos Postales México
- SEPOMEX API
- GeoJSON México
- Georreferenciación
- SQLite
- Arquitectura Hexagonal
- API REST México
- Open Source

## 🤝 Contribuir
¡Las contribuciones son lo que hace a la comunidad de código abierto un lugar increíble! Si tienes una sugerencia para mejorar esto, por favor haz un fork del repositorio y crea un pull request.

Haz un Fork del proyecto

Crea tu rama de característica (git checkout -b feature/CaracteristicaIncreible)

Haz commit de tus cambios (git commit -m 'Añadir CaracteristicaIncreible')

Haz Push a la rama (git push origin feature/CaracteristicaIncreible)

Abre un Pull Request

## 📄 Licencia
Este proyecto está bajo la Licencia BSD-3 - mira el archivo LICENSE para más detalles.

## 🤗 Expresiones de Gratitud
Comparte este proyecto con otros desarrolladores 🗣📢

Invítame una cerveza 🍺 o un café ☕

Da una estrella ⭐ al repositorio si te ha sido útil.

Made with ❤️ by macarthuror - @MacarthurOr - arturo.ortegaro@gmail.com

[contributors-shield]: https://img.shields.io/github/contributors/open-mexico/go-mexpost.svg?style=flat-square
[contributors-url]: https://github.com/open-mexico/go-mexpost/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/open-mexico/go-mexpost.svg?style=flat-square
[forks-url]: https://github.com/open-mexico/go-mexpost/network/members
[stars-shield]: https://img.shields.io/github/stars/open-mexico/go-mexpost.svg?style=flat-square
[stars-url]: https://github.com/open-mexico/go-mexpost/stargazers
[issues-shield]: https://img.shields.io/github/issues/open-mexico/go-mexpost.svg?style=flat-square
[issues-url]: https://github.com/open-mexico/go-mexpost/issues
[license-shield]: https://img.shields.io/github/license/open-mexico/go-mexpost.svg?style=flat-square
[license-url]: https://github.com/open-mexico/go-mexpost/blob/main/LICENSE
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=flat-square&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/macarthuror/