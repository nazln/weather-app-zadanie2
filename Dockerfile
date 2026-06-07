# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

# Certyfikaty SSL potrzebne do połączeń HTTPS z API Open-Meteo
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Osobna warstwa dla zależności — cache nie jest unieważniany przy zmianie kodu
COPY go.mod ./
RUN go mod download

# Kopiowanie kodu źródłowego z kontekstu budowania
COPY main.go ./
COPY templates/ ./templates/

# Statyczny binarny plik — CGO_ENABLED=0 wymagane dla obrazu scratch
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o weather-app main.go

FROM scratch

LABEL org.opencontainers.image.authors="Nazarii Loboda" \
      org.opencontainers.image.title="Aplikacja Pogodowa" \
      org.opencontainers.image.description="Aplikacja webowa pokazująca aktualną pogodę" \
      org.opencontainers.image.version="1.0.0"

# Certyfikaty SSL potrzebne do połączeń HTTPS z API Open-Meteo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/weather-app /weather-app
COPY --from=builder /app/templates/ /templates/

EXPOSE 8089

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/weather-app", "-health-check"]

ENTRYPOINT ["/weather-app"]
