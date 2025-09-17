# Tekton Hub to Artifact Hub Translation Proxy

A Go-based HTTP proxy that translates Tekton Hub API calls to Artifact Hub format, enabling seamless integration between systems expecting Tekton Hub API while using Artifact Hub as the backend.

## Features

- **API Translation**: Converts Tekton Hub API endpoints to Artifact Hub format
- **Catalog Mapping**: Configurable mapping between Tekton Hub and Artifact Hub catalog names
- **Version Conversion**: Handles conversion between simplified semver (0.1) and full semver (0.1.0)
- **Response Format Translation**: Converts Artifact Hub responses to Tekton Hub format
- **Comprehensive Middleware**: Includes CORS, logging, and recovery middleware
- **Health Checks**: Built-in health check endpoint
- **Docker Support**: Containerized deployment ready

## Architecture

```
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
```

### Environment Variables

All configuration can be overridden with environment variables using the `THP_` prefix:

- `THP_SERVER_PORT=8080`
- `THP_ARTIFACTHUB_BASE_URL=https://artifacthub.io`
- `THP_LOGGING_LEVEL=debug`

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

```
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

- **ID-based lookups**: Endpoints requiring specific IDs return `501 Not Implemented` as Artifact Hub doesn't provide direct ID mapping
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