dev:
	go install github.com/air-verse/air@latest
	air

build:
	go build -o go-waf cmd/main.go

dockerize:
	docker build . -t go-waf

docker-dev: dockerize
	docker run -v ./.env:/app/.env -v ./tmp:/storage -p 8080:8080 --rm --cpus="1.0" --memory="1g" go-waf /app/go-waf

save:
	docker save go-waf | gzip > go-waf.tar.gz