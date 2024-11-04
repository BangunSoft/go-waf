dev:
	go install github.com/air-verse/air@latest
	air

build:
	go build -o go-waf cmd/main.go

dockerize:
	docker build . -t go-waf --network=host

docker-dev: dockerize
	docker run -v ./.env:/app/.env -p 8080:8080 --rm go-waf /app/go-waf