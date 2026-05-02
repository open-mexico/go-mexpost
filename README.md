[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<br />
<p align="center">
  <h1 align="center">Go-MexPost 🇲🇽🚀</h1>

  <div align="center">
    <img align="center" src="assets/mex-post.png" alt="MexPost Logo">
  </div>

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
* 🔎 **Búsquedas Flexibles y Optimizadas:** * Por Código Postal exacto o parcial (ej. `067%`).
  * Por Nombre de Colonia exacto o parcial (ignorando mayúsculas y minúsculas).
  * Filtros combinados por Estado y Municipio.
* 🗺️ **Soporte Geoespacial (GeoJSON):** Capacidad de descargar una versión espacial de la base de datos que incluye los polígonos de cada código postal.
* 📍 **Geocodificación (Próximamente):** Búsqueda de colonias mediante coordenadas (Point-in-Polygon) y obtención de centroides geográficos.
* 🏗️ **Arquitectura Hexagonal:** Código desacoplado, testeable y preparado para escalar.

## 🚀 Instalación y Uso

Configurar el proyecto en tu máquina local es increíblemente sencillo gracias a nuestro script de inicialización integrado.

### 1. Clonar el repositorio
```bash
git clone [https://github.com/open-mexico/go-mexpost.git](https://github.com/open-mexico/go-mexpost.git)
cd go-mexpost