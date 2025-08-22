# İletken - HTTP Redirector

**İletken** is a high-performance HTTP redirector application developed with Go 1.24 and fasthttp.

## Features

- ⚡ **High Performance**: Optimized with fasthttp library
- 📝 **YAML Configuration**: Flexible configuration with spf13/viper
- 🔄 **Simple Redirects**: All redirects use 302 (temporary redirect) status code
- 📊 **Structured Logging**: Detailed logging in JSON/Text format
- 🛡️ **Graceful Shutdown**: Safe shutdown with signal handling
- ⚙️ **Easy Configuration**: Simple setup with YAML file
- 🏠 **Default Page**: Built-in index page showing service status
- 🩺 **Health Check**: `/health` endpoint for monitoring

## Installation

```bash
git clone <repository>
cd iletken
go mod download
go build -o iletken
```

## Usage

### 1. Configuration

Edit the `iletken.yml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"

redirects:
  - from: "devops.company.com"
    to: "https://hd.company.com/board/13"
  
  - from: "old.example.com"
    to: "https://new.example.com"

logging:
  level: "info"
  format: "json"
```

### 2. Running the Application

```bash
# With default iletken.yml
./iletken

# With different config file
./iletken -config /path/to/myconfig.yml

# Version information
./iletken -version
```

### 3. Testing

```bash
# Send HTTP request
curl -H "Host: devops.company.com" http://localhost:8080

# Response:
# HTTP/1.1 302 Found
# Location: https://hd.company.com/board/13
```

### 4. Endpoints

- `/` - Default index page with service statistics
- `/health` - Health check endpoint (JSON response)

## Configuration Options

### Server Settings
- `host`: IP address to listen on
- `port`: Port number
- `read_timeout`: Read timeout
- `write_timeout`: Write timeout  
- `idle_timeout`: Idle timeout

### Redirect Rules
- `from`: Source host (domain)
- `to`: Target URL (must be complete URL)

**Note**: All redirects use HTTP 302 (temporary redirect) status code.

### Logging
- `level`: Log level (debug, info, warn, error)
- `format`: Log format (json, text)

## Running with Docker

Create Dockerfile:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o iletken

FROM cgr.dev/chainguard/wolfi-base:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/iletken .
COPY --from=builder /app/iletken.yml .
EXPOSE 8080
CMD ["./iletken"]
```

```bash
docker build -t iletken .
docker run -p 8080:8080 -v $(pwd)/iletken.yml:/app/iletken.yml iletken
```

## Systemd Service

`/etc/systemd/system/iletken.service`:

```ini
[Unit]
Description=İletken HTTP Redirector
After=network.target

[Service]
Type=simple
User=iletken
WorkingDirectory=/opt/iletken
ExecStart=/opt/iletken/iletken -config /opt/iletken/iletken.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable iletken
sudo systemctl start iletken
```

## Performance

High performance thanks to fasthttp usage:
- Low memory allocation
- Fast HTTP parsing
- Zero-copy operations
- Connection pooling

## License

MIT License
