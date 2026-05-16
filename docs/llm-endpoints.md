# LLM Integration Guide for go-mexpost API

This document is an integration contract for AI agents that need to consume go-mexpost endpoints safely and predictably.

If you need contribution rules for AI agents that modify this repository (lint, testing, docs updates), use `/agentes-llm` and `AGENTS.md`.

Source references used:

- docs/endpoints.md
- docs/arquitectura.md
- docs/geocodificacion.md
- docs/configuracion.md

## 1. API Overview

- Base URL: `http://localhost:8080`
- Methods: read-only `GET`
- Core endpoints:
  - `GET /colonias`
  - `GET /municipios`
  - `GET /coordenadas`

## 2. Agent Operating Rules

1. Never call `/colonias` without at least one of: `cp` or `nombre`.
2. For `cp`, enforce digits only and length 3 to 5.
3. For `nombre`, prefer 3 or more chars for better precision.
4. Use `incluir_geo=true` only when geometry is really needed (payload is larger).
5. For `/coordenadas`, always validate latitude/longitude ranges before calling API.
6. On `404`, do not retry immediately with same input; adjust filters first.
7. On `429` or `5xx`, retry with backoff.

## 3. Endpoint Contracts

## 3.1 `GET /colonias`

### Required logic

At least one must be present:

- `cp`
- `nombre`

### Query parameters

- `cp: string`
  - 3 to 5 digits
  - examples: `067`, `0671`, `06700`
- `nombre: string`
  - colonia name, exact or partial
  - case-insensitive
- `municipio_id: string` (optional)
- `incluir_municipio: bool` (optional, default `false`)
- `incluir_geo: bool` (optional, default `false`)
- `solo_geo: bool` (optional, default `false`)
- `limit: int` (optional)
  - default `100`, max `500`
  - if `incluir_geo=true`: default `50`, max `100`
- `pagina: int` (optional, default `1`, min `1`)

### Response shape (success)

```json
{
  "resultados": [
    {
      "codigo": "06700",
      "nombre": "Roma Norte",
      "tipo": "Colonia",
      "ciudad": "Ciudad de Mexico",
      "zona": "Urbano",
      "estado_id": "09",
      "municipio_id": "015",
      "municipio_nombre": "Cuauhtemoc"
    }
  ],
  "total": 42,
  "limit": 10,
  "pagina": 1,
  "total_paginas": 5,
  "pagina_anterior": null,
  "pagina_siguiente": "/colonias?cp=067&limit=10&pagina=2"
}
```

If `incluir_geo=true`, each item may include:

- `geometria` (GeoJSON string)
- `centro_lon` (number)
- `centro_lat` (number)

If `incluir_municipio=true`, each item may include:

- `municipio_nombre` (string), resolved by matching both `municipio_id` and `estado_id`

### Error patterns

- `400` invalid params
- `404` no matches
- `500` internal error
- `429` rate limit exceeded (from architecture docs)

## 3.2 `GET /municipios`

### Required logic

At least one must be present:

- `nombre`
- `estado_id`

### Query parameters

- `nombre: string` (partial, case-insensitive)
- `estado_id: string`

### Response shape (success)

```json
{
  "resultados": [
    {
      "id": "120",
      "nombre": "Zapopan",
      "estado_id": "14"
    }
  ]
}
```

### Error patterns

- `400` invalid params
- `404` no matches

## 3.3 `GET /coordenadas`

Reverse geocoding: finds the colonia that contains a geographic point.

### Query parameters

- `lat: float64` (required, `-90` to `90`)
- `lon: float64` (required, `-180` to `180`)
- `estado_id: string` (optional, recommended for performance)
- `incluir_geo: bool` (optional, default `false`)

### Response shape (success)

```json
{
  "resultado": {
    "codigo": "06700",
    "nombre": "Roma Norte",
    "tipo": "Colonia",
    "ciudad": "Ciudad de Mexico",
    "zona": "Urbano",
    "estado_id": "09",
    "municipio_id": "015"
  }
}
```

Note: `/coordenadas` returns one object in `resultado`, not an array.

### Error patterns

- `400` missing or invalid `lat`/`lon`
- `400` out-of-range coordinates
- `404` point not inside any colonia geometry

## 4. Normalized Types for Client Code

```ts
export type Colonia = {
  codigo: string;
  nombre: string;
  tipo: string;
  ciudad: string;
  zona: string;
  estado_id: string;
  municipio_id: string;
  municipio_nombre?: string;
  geometria?: string;
  centro_lon?: number;
  centro_lat?: number;
};

export type Municipio = {
  id: string;
  nombre: string;
  estado_id: string;
};

export type ColoniasResponse = {
  resultados: Colonia[];
  total: number;
  limit: number;
  pagina: number;
  total_paginas: number;
  pagina_anterior: string | null;
  pagina_siguiente: string | null;
};

export type MunicipiosResponse = {
  resultados: Municipio[];
};

export type CoordenadasResponse = {
  resultado: Colonia;
};

export type ApiError = {
  error: string;
  detalle?: string;
};
```

## 5. Agent Integration Workflow (Recommended)

1. If user gives free text place + state, call `/municipios` first to resolve `estado_id` and optional `municipio_id`.
2. For postal search, call `/colonias` with the narrowest valid filter set.
3. If user gives coordinates, call `/coordenadas` directly.
4. Only request geometry when map rendering is needed.
5. For lists, support pagination using `pagina_siguiente` and `pagina_anterior`.

## 6. Retry and Fallback Policy

- `400`: do not retry. Repair input and ask user clarification.
- `404`: do not blind-retry. Broaden query (`cp` prefix, partial `nombre`, remove extra filter).
- `429`: retry with exponential backoff and jitter.
- `500`: retry up to 2 times with short exponential backoff.

Suggested backoff: `300ms`, `900ms`, `1800ms` (+ random jitter 0-250ms).

## 7. Copy/Paste Starter Pack (Any LLM)

This section is intentionally designed to be copied as-is into your preferred LLM workflow.

### One-click buttons

<CopyForLLM
title="System Prompt - Strict Integration Mode"
language="markdown"
:content="`You are an API integration assistant for go-mexpost.

Your job is to build safe, valid requests against this API only:

- Base URL: http://localhost:8080
- Endpoints: GET /colonias, GET /municipios, GET /coordenadas

Validation rules:

- /colonias: require at least one of cp or nombre.
- cp must be digits only, length 3 to 5.
- pagina must be integer >= 1.
- lat must be in [-90, 90], lon in [-180, 180].

Behavior rules:

- Use incluir_geo=true only when geometry is required.
- On 400: do not retry; repair input.
- On 404: do not blind retry; broaden or adjust filters.
- On 429 or 5xx: retry with exponential backoff and jitter.
- Preserve API response field names exactly.

Output style:

- Be concise.
- If input is ambiguous, ask one clear follow-up question.
- When calling the API, show endpoint and query params explicitly.`"
  />

<CopyForLLM
title="Developer Prompt - Drop-in Template"
language="markdown"
:content="`Integrate go-mexpost API requests from user intent.

Request policy:

1. Parse user intent and decide one endpoint:
   - postal or colonia search -> /colonias
   - municipality lookup -> /municipios
   - coordinate reverse geocoding -> /coordenadas
2. Validate all params before building the request.
3. Build only GET requests.
4. Keep filters as narrow as possible first.
5. Add incluir_geo=true only when map/shape output is explicitly requested.

Failure policy:

- 400: explain which input is invalid and how to fix it.
- 404: suggest one broader query.
- 429/5xx: retry with backoff (300ms, 900ms, 1800ms + jitter).

Response policy:

- Keep answer structured.
- Include endpoint used, normalized parameters, and key results.
- Never invent fields not returned by the API.`"
  />

<CopyForLLM
title="cURL Smoke Test - One Paste"
language="bash"
:content="`BASE_URL=\"\${BASE_URL:-http://localhost:8080}\"

echo \"[1/7] /colonias by cp\"
curl -fsS \"$BASE_URL/colonias?cp=067&limit=20&pagina=1\" | head -c 500; echo

echo \"[2/7] /colonias by name + municipio\"
curl -fsS \"$BASE_URL/colonias?nombre=Roma&municipio_id=015\" | head -c 500; echo

echo \"[3/7] /colonias geo payload\"
curl -fsS \"$BASE_URL/colonias?cp=067&incluir_geo=true&solo_geo=true\" | head -c 500; echo

echo \"[4/7] /municipios by name\"
curl -fsS \"$BASE_URL/municipios?nombre=Zapopan\" | head -c 500; echo

echo \"[5/7] /municipios by state\"
curl -fsS \"$BASE_URL/municipios?estado_id=14\" | head -c 500; echo

echo \"[6/7] /coordenadas basic\"
curl -fsS \"$BASE_URL/coordenadas?lat=19.4181&lon=-99.1634\" | head -c 500; echo

echo \"[7/7] /coordenadas with state + geo\"
curl -fsS \"$BASE_URL/coordenadas?lat=19.4181&lon=-99.1634&estado_id=09&incluir_geo=true\" | head -c 500; echo`"
/>

### 7.1 System Prompt (Strict Integration Mode)

Copy this block directly as your **system prompt**:

```md
You are an API integration assistant for go-mexpost.

Your job is to build safe, valid requests against this API only:

- Base URL: http://localhost:8080
- Endpoints: GET /colonias, GET /municipios, GET /coordenadas

Validation rules:

- /colonias: require at least one of cp or nombre.
- cp must be digits only, length 3 to 5.
- pagina must be integer >= 1.
- lat must be in [-90, 90], lon in [-180, 180].

Behavior rules:

- Use incluir_geo=true only when geometry is required.
- On 400: do not retry; repair input.
- On 404: do not blind retry; broaden or adjust filters.
- On 429 or 5xx: retry with exponential backoff and jitter.
- Preserve API response field names exactly.

Output style:

- Be concise.
- If input is ambiguous, ask one clear follow-up question.
- When calling the API, show endpoint and query params explicitly.
```

### 7.2 Developer Prompt (Drop-in Template)

Copy this block as your **developer prompt** (or tool instruction layer):

```md
Integrate go-mexpost API requests from user intent.

Request policy:

1. Parse user intent and decide one endpoint:

- postal or colonia search -> /colonias
- municipality lookup -> /municipios
- coordinate reverse geocoding -> /coordenadas

2. Validate all params before building the request.
3. Build only GET requests.
4. Keep filters as narrow as possible first.
5. Add incluir_geo=true only when map/shape output is explicitly requested.

Failure policy:

- 400: explain which input is invalid and how to fix it.
- 404: suggest one broader query.
- 429/5xx: retry with backoff (300ms, 900ms, 1800ms + jitter).

Response policy:

- Keep answer structured.
- Include endpoint used, normalized parameters, and key results.
- Never invent fields not returned by the API.
```

### 7.3 One-Paste cURL Smoke Test

Use this block to quickly validate your local integration:

```bash
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "[1/7] /colonias by cp"
curl -fsS "$BASE_URL/colonias?cp=067&limit=20&pagina=1" | head -c 500; echo

echo "[2/7] /colonias by name + municipio"
curl -fsS "$BASE_URL/colonias?nombre=Roma&municipio_id=015" | head -c 500; echo

echo "[3/7] /colonias geo payload"
curl -fsS "$BASE_URL/colonias?cp=067&incluir_geo=true&solo_geo=true" | head -c 500; echo

echo "[4/7] /municipios by name"
curl -fsS "$BASE_URL/municipios?nombre=Zapopan" | head -c 500; echo

echo "[5/7] /municipios by state"
curl -fsS "$BASE_URL/municipios?estado_id=14" | head -c 500; echo

echo "[6/7] /coordenadas basic"
curl -fsS "$BASE_URL/coordenadas?lat=19.4181&lon=-99.1634" | head -c 500; echo

echo "[7/7] /coordenadas with state + geo"
curl -fsS "$BASE_URL/coordenadas?lat=19.4181&lon=-99.1634&estado_id=09&incluir_geo=true" | head -c 500; echo
```

### 7.4 Quick User Prompts (for manual testing)

Copy any of these into your LLM chat to verify behavior:

- "Find colonias for postal code 06700, page 1, limit 20."
- "Search colonia Roma in municipio 015 and include geometry."
- "List municipios in estado 14."
- "Given lat 19.4181 and lon -99.1634, find the colonia."

## 8. Minimal Prompt Block (Short Version)

If you need a smaller prompt, use this as system or developer instructions:

```md
You are integrating go-mexpost API.

Rules:

- Base URL is http://localhost:8080
- Use only GET endpoints: /colonias, /municipios, /coordenadas
- Validate inputs before calling API:
  - cp: 3-5 digits
  - pagina: integer >= 1
  - lat in [-90, 90], lon in [-180, 180]
- Request incluir_geo=true only when geometry is necessary.
- Handle errors:
  - 400: fix parameters, no retry
  - 404: adjust filters, no blind retry
  - 429 and 5xx: retry with backoff
- Return concise structured results and preserve response fields.
```

## 9. Notes for Accuracy

- Search behavior in docs indicates partial matching for text values with 3+ chars.
- In practice, `cp` also has strict validation (3 to 5 digits), so shorter `cp` values should be treated as invalid.
- Reverse geocoding depends on geometry data in `mapa.db`. If DB was installed with `-geo=false`, `/coordenadas` can return `404` consistently.

## 10. Integration Checklist for Developers

Before shipping your LLM integration:

1. Confirm `BASE_URL` is configurable by environment.
2. Enforce parameter validation before API calls.
3. Implement `400/404/429/5xx` behavior exactly as documented.
4. Add retry jitter to avoid synchronized bursts.
5. Keep geometry optional to control payload size and latency.
6. Log endpoint + params (without sensitive data) for debugging.
