# ==========================================
# Etapa 1: Construcción (Builder)
# ==========================================
FROM golang:alpine AS builder

WORKDIR /app

# Parámetro para decidir si incluimos geometrías en la DB (default: false)
ARG INCLUDE_GEO="false"

# Primero copiamos los archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente
COPY . .

# Generamos la base de datos de manera autónoma usando el script ETL interno
RUN go run ./cmd/setup/main.go -geo=${INCLUDE_GEO}

# Compilamos el binario de forma estática
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/go-mexpost-api ./cmd/api/main.go

# ==========================================
# Etapa 2: Producción (Final)
# ==========================================
FROM alpine:latest

# Añadimos certificados de seguridad
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiamos el binario compilado desde la etapa 1
COPY --from=builder /app/go-mexpost-api .

# Copiamos la base de datos generada desde la etapa 1
COPY --from=builder /app/mapa.db .

# Exponemos el puerto
EXPOSE 8080

# Comando de arranque
CMD ["./go-mexpost-api"]
