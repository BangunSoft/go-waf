FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app
COPY . /app

RUN go build -o go-waf cmd/main.go

# =================
FROM alpine:3.20 

WORKDIR /app
ENV PATH="$PATH:/app" \
    TZ=Asia/Jakarta
RUN apk add --no-cache tzdata \
    && mkdir -p /app/cache
COPY --from=builder /app/go-waf /app/go-waf
COPY devices /app/devices
COPY views /app/views
COPY .env-example /app/.env-example

EXPOSE 8080
CMD [ "go-waf" ]