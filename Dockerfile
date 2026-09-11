# Etapa 1: Construcción
FROM golang:1.21-alpine AS builder

# Instalar git para descargar dependencias si es necesario
RUN apk update && apk add --no-cache git tzdata ca-certificates

WORKDIR /app

# Copiar go.mod y go.sum primero para aprovechar la caché de Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar la aplicación estáticamente (sin dependencias dinámicas)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o verifyDomain .

# Etapa 2: Imagen final (ultra ligera)
FROM alpine:latest

# Añadir certificados para que las peticiones HTTPS externas funcionen
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiar el ejecutable desde la etapa de construcción
COPY --from=builder /app/verifyDomain .

# Exponer el puerto (opcional, documentativo)
EXPOSE 8080

# Ejecutar el binario
CMD ["./verifyDomain"]
