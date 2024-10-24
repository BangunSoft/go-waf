dev:
	go install github.com/air-verse/air@latest
	air

build:
	go build -o go-waf cmd/main.go

dockerize:
	docker build . -t go-waf --network=host