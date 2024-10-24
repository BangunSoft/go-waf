# go-waf

`go-waf` is a web application firewall (WAF) built using Go and the Gin framework. It provides features such as rate limiting, caching, and reverse proxying to enhance the security and performance of web applications.

## Features

- **Rate Limiting**: Control the number of requests a client can make in a given time period.
- **Caching**: Cache responses to improve performance and reduce load on backend services.
- **Reverse Proxy**: Forward requests to backend services while handling SSL termination and other proxy-related tasks.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/jahrulnr/go-waf.git
   cd go-waf
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the application:

   ```bash
   go build -o go-waf cmd/main.go
   ```

### Running the Application

To run the application in development mode:

```bash
make dev
```

To build and run the application using Docker:

```bash
make dockerize
docker run -p 8080:8080 go-waf
```

### Configuration

The application can be configured using environment variables or a `.env` file. Refer to `config/config.go` for available configuration options.

### Usage

- **Rate Limiting**: Configure rate limiting settings in the environment variables or `.env` file.
- **Caching**: Enable caching and choose a cache driver (memory, file, or Redis) in the configuration.
- **Reverse Proxy**: Set the `HOST_DESTINATION` to the backend service URL.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Logrus](https://github.com/sirupsen/logrus) for logging
- [Redis](https://github.com/redis/go-redis) for caching