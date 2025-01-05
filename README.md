# go-waf

`go-waf` is a web application firewall (WAF) developed in Go using the Gin framework. It provides robust security features to protect web applications from various threats while ensuring high performance and scalability.

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the Application](#running-the-application)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Reverse Proxy](#reverse-proxy)
  - [Web Application Firewall (WAF)](#web-application-firewall-waf)
  - [Rate Limiting](#rate-limiting)
  - [Cache Configuration](#cache-configuration)
  - [Clearing Cache](#clearing-cache)
- [License](#license)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)

## Features

- **Rate Limiting**: Manage and restrict the number of requests a client can make within a specified time frame to prevent abuse.
- **SQL Injection Detection**: Identify and block potential SQL injection attacks to safeguard your database.
- **XSS Injection Detection**: Detect and mitigate cross-site scripting (XSS) attacks to protect user data and maintain application integrity.
- **Caching**: Improve response times and reduce backend load by caching frequently accessed data.
- **Reverse Proxy**: Seamlessly forward requests to backend services, handling SSL termination and other proxy-related tasks.

## Getting Started

### Prerequisites

- Go 1.23 or later for development
- Docker for containerized deployment

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

3. Copy the example environment file:

   ```bash
   cp .env-example .env
   ```

   Update the values in the `.env` file as needed.

4. Build the application:

   ```bash
   go build -o go-waf cmd/main.go
   ```

### Running the Application

To run the application in development mode with live reloading:

```bash
make dev
```

To build and run the application using Docker:

```bash
make docker-dev
```

### Configuration

The application can be configured using environment variables or a `.env` file. For a comprehensive list of available configuration options, refer to `config/config.go`.

### Usage
#### **Reverse Proxy**
Configure the reverse proxy settings in your `.env` file:
- `ADDR=:8080`: The port on which the service will listen.
- `HOST=bangunsoft.com`: The domain used for masking the destination.
- `HOST_DESTINATION=http://my-app:3000`: The actual backend service URL.
The application will fetch data from the backend service and replace the `HOST_DESTINATION` domain with the `HOST` domain in the response. This is particularly useful for local development or docker hostname. For example:
- If you set `HOST=bangunsoft.com` and `HOST_DESTINATION=http://my-app:3000`, the application will replace `http://my-app:3000` with `http://bangunsoft.com` in the response.

#### **Web Application Firewall (WAF)**
Enable the WAF by setting `USE_WAF=true` in your `.env` file.<br/>
Configure the WAF settings:
- `WAF_CONFIG=config/keywords.yml`: Specify the path to the WAF configuration file.
- `WAF_PROTECT_HEADER=true`: Enable protection for HTTP headers.
- `WAF_PROTECT_BODY=true`: Enable protection for the body of requests.

#### **Rate Limiting**
  Enable rate limiting by setting `USE_RATELIMIT=true` in your `.env` file.
  Configure the rate limiting settings:
  - `RATELIMIT_SECOND=1`: The time window for rate limiting, in seconds.
  - `RATELIMIT_MAX=50`: The maximum number of requests allowed within the specified time window. For example, with the above settings, a client can make up to 50 requests per second.

#### **Cache Configuration**
  - `USE_CACHE=true`: Enable caching.
  - `CACHE_TTL=3600`: Set the time-to-live for cached items (in seconds).
  - `CACHE_DRIVER=file`: Specify the cache driver to use.
  - `CACHE_REMOVE_METHOD=ban`: Method to remove cached items.
  - `CACHE_REMOVE_ALLOW_IP=127.0.0.1,::1,127.0.0.0/8`: IP addresses allowed to remove cache items.

#### **Clearing Cache**
  - To delete a specific cache entry, use the following command:
    ```bash 
    curl localhost:8080/blog -X BAN
    ```
  - To bulk delete cache entries that match a prefix, use:
    ```bash 
    curl localhost:8080/blog?is_prefix=true -X BAN
    ```
    This command will remove all cache entries that start with `/blog`.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Logrus](https://github.com/sirupsen/logrus) for logging
- [Redis](https://github.com/redis/go-redis/v9) for caching
- [gin-rate-limit](https://github.com/JGLTechnologies/gin-rate-limit) for rate limiting
- [gzip](https://github.com/nanmu42/gzip) for Gzip compression
- [libinjection-go](https://github.com/corazawaf/libinjection-go) for injection detection
- Etc.