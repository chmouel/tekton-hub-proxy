[![Claude](https://img.shields.io/badge/Claude-D97757?logo=claude&logoColor=fff)](#) [![Google Gemini](https://img.shields.io/badge/Google%20Gemini-886FBF?logo=googlegemini&logoColor=fff)](#)

# Tekton Hub to Artifact Hub Translation Proxy

A Go-based HTTP proxy that translates Tekton Hub API calls to Artifact Hub
format, enabling seamless integration between systems expecting Tekton Hub API
while using Artifact Hub as the backend.

## Features

- **API Translation**: Converts Tekton Hub API endpoints to Artifact Hub format
- **Catalog Mapping**: Configurable mapping between Tekton Hub and Artifact Hub
  catalog names
- **Version Conversion**: Handles conversion between simplified semver (0.1) and
  full semver (0.1.0)
- **Response Format Translation**: Converts Artifact Hub responses to Tekton Hub
  format
- **Comprehensive Middleware**: Includes CORS, logging, and recovery middleware
- **Health Checks**: Built-in health check endpoint
- **Docker Support**: Containerized deployment ready

## Demo

[![Tekton Hub Proxy Demo](https://img.youtube.com/vi/lzVMj7fKnUA/0.jpg)](https://www.youtube.com/watch?v=lzVMj7fKnUA)

## Architecture

```ascii
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Tekton Hub    │    │  Translation    │    │  Artifact Hub   │
│     Client      │───▶│     Proxy       │───▶│      API        │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## API Endpoints

The proxy implements the complete Tekton Hub API:

### Catalog Endpoints

- `GET /v1/catalogs` - List available catalogs

### Resource Endpoints

- `GET /v1/resource/{catalog}/{kind}/{name}` - Get resource details
- `GET /v1/resource/{catalog}/{kind}/{name}/{version}` - Get specific version
- `GET /v1/resource/{catalog}/{kind}/{name}/{version}/yaml` - Get YAML content
- `GET /v1/resource/{catalog}/{kind}/{name}/{version}/readme` - Get README
- `GET /v1/resource/{catalog}/{kind}/{name}/raw` - Get latest raw YAML
- `GET /v1/resource/{catalog}/{kind}/{name}/{version}/raw` - Get raw YAML for version

### Query Endpoints

- `GET /v1/resources` - List all resources
- `GET /v1/query` - Search resources with filters

### Health

- `GET /health` - Health check endpoint

## Configuration

Configuration is managed via YAML files and environment variables:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

artifacthub:
  base_url: "https://artifacthub.io"
  timeout: 30s
  max_retries: 3

catalog_mappings:
  - tekton_hub: "tekton"
    artifact_hub: "tekton-catalog-tasks"
  - tekton_hub: "tekton-community"
    artifact_hub: "tekton-catalog-community"

logging:
  level: "info"
  format: "json"

landing_page:
  enabled: true  # Set to false to disable the landing page
```

### Environment Variables

All configuration can be overridden with environment variables using the `THP_` prefix:

- `THP_SERVER_PORT=8080`
- `THP_SERVER_HOST=192.168.1.100`
- `THP_ARTIFACTHUB_BASE_URL=https://artifacthub.io`
- `THP_LOGGING_LEVEL=debug`
- `THP_LANDING_PAGE_ENABLED=false`

### Command Line Flags

The application supports several command line flags:

```bash
./bin/tekton-hub-proxy [options]

Options:
  --config string               Path to config file
  --port int                   Server port (overrides config)
  --bind string                Bind address (overrides config)
  --debug                      Enable debug logging
  --disable-landing-page       Disable the landing page at root path (/)
  --help                       Show help message
```

**Examples**:
```bash
# Start with custom port and disabled landing page
./bin/tekton-hub-proxy --port 9090 --disable-landing-page

# Start with debug logging and custom config
./bin/tekton-hub-proxy --debug --config /path/to/config.yaml

# Show help
./bin/tekton-hub-proxy --help
```

## Quick Start

### Local Development

1. **Clone and setup**:

   ```bash
   git clone <repository>
   cd tekton-hub-proxy
   go mod download
   ```

2. **Run the server**:

   ```bash
   go run cmd/server/main.go

   # Or with custom options
   go run cmd/server/main.go -debug -port 9090 -bind 192.168.1.100
   ```

3. **Test the API**:

   ```bash
   curl http://localhost:8080/v1/catalogs
   curl http://localhost:8080/v1/resource/tekton/task/buildah
   ```

### Docker Deployment

1. **Build the image**:

   ```bash
   docker build -t tekton-hub-proxy .
   ```

2. **Run the container**:

   ```bash
   docker run -p 8080:8080 \
     -e THP_LOGGING_LEVEL=debug \
     tekton-hub-proxy
   ```

3. **With custom configuration**:

   ```bash
   docker run -p 8080:8080 \
     -v $(pwd)/custom-config.yaml:/root/configs/config.yaml \
     tekton-hub-proxy
   ```

## Reverse Proxy Deployment

For production deployments, it's recommended to deploy the proxy behind a reverse proxy like Nginx or Caddy for SSL termination, load balancing, and aggressive caching.

**Note**: The configurations below include the landing page route (`/`). If you disable the landing page using `--disable-landing-page`, remove the corresponding location/handle block from your reverse proxy configuration.

### Nginx Deployment

Create an Nginx configuration file (`/etc/nginx/sites-available/tekton-hub-proxy`):

```nginx
upstream tekton_hub_proxy {
    server 127.0.0.1:8080;
    # Add more servers for load balancing
    # server 127.0.0.1:8081;
}

# Cache configuration
proxy_cache_path /var/cache/nginx/tekton_hub levels=1:2 keys_zone=tekton_hub:10m max_size=1g inactive=60m use_temp_path=off;

server {
    listen 443 ssl http2;
    server_name tekton-hub.example.com;

    # SSL configuration
    ssl_certificate /path/to/your/cert.pem;
    ssl_certificate_key /path/to/your/key.pem;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Gzip compression
    gzip on;
    gzip_types text/plain application/json application/yaml text/yaml text/html;

    # Landing page (optional - can be disabled)
    location = / {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Short cache for landing page
        proxy_cache tekton_hub;
        proxy_cache_valid 200 1h;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location /health {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Short cache for health checks
        proxy_cache tekton_hub;
        proxy_cache_valid 200 30s;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location /v1/catalogs {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Aggressive caching for catalog list (changes infrequently)
        proxy_cache tekton_hub;
        proxy_cache_valid 200 24h;
        proxy_cache_valid 404 1h;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location ~* /v1/resource/.+/yaml$ {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Very aggressive caching for YAML content (immutable per version)
        proxy_cache tekton_hub;
        proxy_cache_valid 200 7d;
        proxy_cache_valid 404 1h;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location ~* /v1/resource/.+/readme$ {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Long caching for README content
        proxy_cache tekton_hub;
        proxy_cache_valid 200 24h;
        proxy_cache_valid 404 1h;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location ~* /v1/resource {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Medium caching for resource metadata
        proxy_cache tekton_hub;
        proxy_cache_valid 200 6h;
        proxy_cache_valid 404 30m;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location /v1/ {
        proxy_pass http://tekton_hub_proxy;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Default caching for other API endpoints
        proxy_cache tekton_hub;
        proxy_cache_valid 200 1h;
        proxy_cache_valid 404 10m;
        proxy_cache_use_stale error timeout invalid_header updating;
        add_header X-Cache-Status $upstream_cache_status;
    }
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name tekton-hub.example.com;
    return 301 https://$server_name$request_uri;
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/tekton-hub-proxy /etc/nginx/sites-enabled/
sudo mkdir -p /var/cache/nginx/tekton_hub
sudo chown www-data:www-data /var/cache/nginx/tekton_hub
sudo nginx -t && sudo systemctl reload nginx
```

### Caddy Deployment

Create a Caddyfile:

```caddy
tekton-hub.example.com {
    # Enable compression
    encode gzip

    # Landing page (optional - can be disabled with --disable-landing-page)
    handle / {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=3600, stale-while-revalidate=600"
    }

    # Health check with short cache
    handle /health {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=30"
    }

    # Catalog endpoints with long cache
    handle /v1/catalogs {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=86400, stale-while-revalidate=3600"
    }

    # YAML content with very long cache (immutable)
    handle_path /v1/resource/*/*/yaml {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=604800, immutable"
    }

    # Raw YAML content with very long cache
    handle_path /v1/resource/*/raw {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=604800, immutable"
    }

    # README content with medium cache
    handle_path /v1/resource/*/*/readme {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=86400, stale-while-revalidate=3600"
    }

    # Resource metadata with medium cache
    handle_path /v1/resource/* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=21600, stale-while-revalidate=3600"
    }

    # Query and resources endpoints with shorter cache
    handle /v1/query* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=3600, stale-while-revalidate=600"
    }

    handle /v1/resources* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=3600, stale-while-revalidate=600"
    }

    # Default for other API endpoints
    handle /v1/* {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=3600, stale-while-revalidate=600"
    }

    # Security headers
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
        Referrer-Policy strict-origin-when-cross-origin
    }
}
```

### Docker Compose Examples

#### Nginx + Proxy

```yaml
# docker-compose.nginx.yml
version: '3.8'

services:
  tekton-hub-proxy:
    image: tekton-hub-proxy:latest
    container_name: tekton-hub-proxy
    environment:
      - THP_SERVER_HOST=0.0.0.0
      - THP_SERVER_PORT=8080
      - THP_LOGGING_LEVEL=info
    networks:
      - proxy-network
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: nginx-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
      - ./ssl:/etc/ssl/certs
      - nginx-cache:/var/cache/nginx
    depends_on:
      - tekton-hub-proxy
    networks:
      - proxy-network
    restart: unless-stopped

volumes:
  nginx-cache:

networks:
  proxy-network:
    driver: bridge
```

#### Caddy + Proxy

```yaml
# docker-compose.caddy.yml
version: '3.8'

services:
  tekton-hub-proxy:
    image: tekton-hub-proxy:latest
    container_name: tekton-hub-proxy
    environment:
      - THP_SERVER_HOST=0.0.0.0
      - THP_SERVER_PORT=8080
      - THP_LOGGING_LEVEL=info
    networks:
      - proxy-network
    restart: unless-stopped

  caddy:
    image: caddy:alpine
    container_name: caddy-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
      - caddy-config:/config
    depends_on:
      - tekton-hub-proxy
    networks:
      - proxy-network
    restart: unless-stopped

volumes:
  caddy-data:
  caddy-config:

networks:
  proxy-network:
    driver: bridge
```

### Systemd Container Service

For production deployments on systems with systemd, you can run the proxy as a systemd container service using Podman or Docker.

#### Using Podman (Recommended)

Create a systemd service file (`/etc/systemd/system/tekton-hub-proxy.service`):

```ini
[Unit]
Description=Tekton Hub Proxy Container
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
NotifyAccess=all
Restart=always
RestartSec=5
TimeoutStartSec=900
TimeoutStopSec=120
ExecStartPre=-/usr/bin/podman stop tekton-hub-proxy
ExecStartPre=-/usr/bin/podman rm tekton-hub-proxy
ExecStart=/usr/bin/podman run --rm --name tekton-hub-proxy \
    --publish 8080:8080 \
    --env THP_LOGGING_LEVEL=info \
    --env THP_SERVER_HOST=0.0.0.0 \
    --env THP_SERVER_PORT=8080 \
    --volume /etc/tekton-hub-proxy:/root/configs:ro,Z \
    --conmon-pidfile=%t/%n.pid \
    --cidfile=%t/%n.cid \
    --cgroups=no-conmon \
    --sdnotify=conmon \
    ghcr.io/your-org/tekton-hub-proxy:latest
ExecStop=/usr/bin/podman stop --ignore --cidfile=%t/%n.cid
ExecStopPost=-/usr/bin/podman rm --force --ignore --cidfile=%t/%n.cid
PIDFile=%t/%n.pid
KillMode=mixed
SyslogIdentifier=tekton-hub-proxy

[Install]
WantedBy=multi-user.target
```

#### Using Docker

Create a systemd service file (`/etc/systemd/system/tekton-hub-proxy.service`):

```ini
[Unit]
Description=Tekton Hub Proxy Container
After=docker.service
Requires=docker.service

[Service]
Type=simple
Restart=always
RestartSec=5
TimeoutStartSec=900
TimeoutStopSec=120
ExecStartPre=-/usr/bin/docker stop tekton-hub-proxy
ExecStartPre=-/usr/bin/docker rm tekton-hub-proxy
ExecStart=/usr/bin/docker run --rm --name tekton-hub-proxy \
    --publish 8080:8080 \
    --env THP_LOGGING_LEVEL=info \
    --env THP_SERVER_HOST=0.0.0.0 \
    --env THP_SERVER_PORT=8080 \
    --volume /etc/tekton-hub-proxy:/root/configs:ro \
    ghcr.io/your-org/tekton-hub-proxy:latest
ExecStop=/usr/bin/docker stop tekton-hub-proxy
SyslogIdentifier=tekton-hub-proxy
User=root

[Install]
WantedBy=multi-user.target
```

#### Setup and Configuration

1. **Create configuration directory**:

   ```bash
   sudo mkdir -p /etc/tekton-hub-proxy
   sudo chown root:root /etc/tekton-hub-proxy
   sudo chmod 755 /etc/tekton-hub-proxy
   ```

2. **Create configuration file** (`/etc/tekton-hub-proxy/config.yaml`):

   ```yaml
   server:
     port: 8080
     host: "0.0.0.0"

   artifacthub:
     base_url: "https://artifacthub.io"
     timeout: 30s
     max_retries: 3

   catalog_mappings:
     - tekton_hub: "tekton"
       artifact_hub: "tekton-catalog-tasks"
     - tekton_hub: "tekton-community"
       artifact_hub: "tekton-catalog-community"

   logging:
     level: "info"
     format: "json"
   ```

3. **Enable and start the service**:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable tekton-hub-proxy.service
   sudo systemctl start tekton-hub-proxy.service
   ```

4. **Check service status**:

   ```bash
   sudo systemctl status tekton-hub-proxy.service
   sudo journalctl -u tekton-hub-proxy.service -f
   ```

#### Service Management Commands

```bash
# Start the service
sudo systemctl start tekton-hub-proxy

# Stop the service
sudo systemctl stop tekton-hub-proxy

# Restart the service
sudo systemctl restart tekton-hub-proxy

# Enable auto-start on boot
sudo systemctl enable tekton-hub-proxy

# Disable auto-start on boot
sudo systemctl disable tekton-hub-proxy

# View service logs
sudo journalctl -u tekton-hub-proxy -f

# View service status
sudo systemctl status tekton-hub-proxy
```

#### User Service (Rootless)

For running as a user service (rootless containers):

1. **Create user service directory**:

   ```bash
   mkdir -p ~/.config/systemd/user
   ```

2. **Create user service file** (`~/.config/systemd/user/tekton-hub-proxy.service`):

   ```ini
   [Unit]
   Description=Tekton Hub Proxy Container (User)
   After=network-online.target
   Wants=network-online.target

   [Service]
   Type=notify
   NotifyAccess=all
   Restart=always
   RestartSec=5
   ExecStartPre=-/usr/bin/podman stop tekton-hub-proxy
   ExecStartPre=-/usr/bin/podman rm tekton-hub-proxy
   ExecStart=/usr/bin/podman run --rm --name tekton-hub-proxy \
       --publish 8080:8080 \
       --env THP_LOGGING_LEVEL=info \
       --volume %h/.config/tekton-hub-proxy:/root/configs:ro,Z \
       --conmon-pidfile=%t/%n.pid \
       --cidfile=%t/%n.cid \
       --cgroups=no-conmon \
       --sdnotify=conmon \
       ghcr.io/your-org/tekton-hub-proxy:latest
   ExecStop=/usr/bin/podman stop --ignore --cidfile=%t/%n.cid
   ExecStopPost=-/usr/bin/podman rm --force --ignore --cidfile=%t/%n.cid
   PIDFile=%t/%n.pid
   KillMode=mixed

   [Install]
   WantedBy=default.target
   ```

3. **Manage user service**:

   ```bash
   # Reload user services
   systemctl --user daemon-reload

   # Enable and start
   systemctl --user enable --now tekton-hub-proxy.service

   # Check status
   systemctl --user status tekton-hub-proxy.service

   # View logs
   journalctl --user -u tekton-hub-proxy.service -f
   ```

### Cache Strategy Explanation

The caching strategy is optimized based on content mutability:

- **Health checks** (`/health`): 30 seconds - frequent checks needed
- **Catalogs** (`/v1/catalogs`): 24 hours - catalog list changes infrequently
- **YAML content** (`/v1/resource/.../yaml`): 7 days - immutable per version
- **README content** (`/v1/resource/.../readme`): 24 hours - documentation updates occasionally
- **Resource metadata** (`/v1/resource/...`): 6 hours - metadata may change but not frequently
- **Query results** (`/v1/query`, `/v1/resources`): 1 hour - search results can change with new packages

## Translation Details

### Catalog Name Mapping

The proxy translates catalog names between the two systems:

- **Tekton Hub**: `tekton` → **Artifact Hub**: `tekton-catalog-tasks`
- **Tekton Hub**: `tekton-community` → **Artifact Hub**: `tekton-catalog-community`

### Version Format Conversion

- **Tekton Hub** uses simplified semver: `0.1`, `0.2`, `1.0`
- **Artifact Hub** uses full semver: `0.1.0`, `0.2.0`, `1.0.0`

The proxy automatically converts between these formats.

### Response Format Translation

**Tekton Hub YAML Response**:

```json
{
  "data": {
    "yaml": "apiVersion: tekton.dev/v1..."
  }
}
```

**Artifact Hub Package Response**:

```json
{
  "data": {
    "manifestRaw": "apiVersion: tekton.dev/v1..."
  }
}
```

## Development

### Project Structure

```console
tekton-hub-proxy/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── client/          # Artifact Hub API client
│   ├── translator/      # Translation logic
│   └── models/          # Data models
├── configs/             # Configuration files
└── Dockerfile           # Container definition
```

### Adding New Catalog Mappings

Edit `configs/config.yaml`:

```yaml
catalog_mappings:
  - tekton_hub: "tekton"
    artifact_hub: "tekton-catalog-tasks"
  - tekton_hub: "my-catalog"
    artifact_hub: "my-artifacthub-catalog"
```

### Logging

The proxy uses structured logging with configurable levels:

```bash
# Set log level
export THP_LOGGING_LEVEL=debug

# Set log format
export THP_LOGGING_FORMAT=text
```

## Monitoring

### Health Checks

The `/health` endpoint returns:

```json
{
  "status": "healthy"
}
```

### Metrics

All requests are logged with:

- Method and path
- Status code
- Response time
- Remote address
- User agent

## Limitations

- **ID-based lookups**: Endpoints requiring specific IDs return `501 Not  
  Implemented` as Artifact Hub doesn't provide direct ID mapping
- **Category/Tag mapping**: Basic keyword-based mapping is used
- **Platform support**: Defaults to `linux/amd64`
- **Rating**: Uses default rating as Artifact Hub doesn't provide this metric

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Apache License 2.0 - see LICENSE file for details.
