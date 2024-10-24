FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app
COPY . /app

RUN go build -o go-waf cmd/main.go

# =================
FROM alpine:3.20 

WORKDIR /app
ENV PATH="$PATH:/app" \
    TZ=Asia/Jakarta
RUN mkdir -p /app/cache
COPY --from=builder /app/go-waf /app/go-waf
COPY .env-example /app/.env-example
RUN apk add --no-cache tzdata

EXPOSE 8080
CMD [ "go-waf" ]