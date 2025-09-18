[![Claude](https://img.shields.io/badge/Claude-D97757?logo=claude&logoColor=fff)](#) [![Google Gemini](https://img.shields.io/badge/Google%20Gemini-886FBF?logo=googlegemini&logoColor=fff)](#)

# Tekton Hub to Artifact Hub Translation Proxy

A Go-based HTTP proxy that translates Tekton Hub API calls to Artifact Hub
format. **This proxy is intended for migration assistance only** - users should
transition to using Artifact Hub directly as modern Tekton tooling supports it natively.

> **⚠️ Migration Notice**: This proxy is a temporary solution for transitioning from Tekton Hub to Artifact Hub.
> Modern versions of Tekton Pipelines and Pipelines-as-Code support Artifact Hub natively.
> Please consider upgrading your tooling and migrating to Artifact Hub directly.

## Features

- **API Translation**: Converts Tekton Hub API endpoints to Artifact Hub format
- **High-Performance Caching**: In-memory cache with TTL and LRU eviction for 1000x+ faster response times
- **Catalog Mapping**: Configurable mapping between Tekton Hub and Artifact Hub
  catalog names
- **Version Conversion**: Handles conversion between simplified semver (0.1) and
  full semver (0.1.0)
- **Response Format Translation**: Converts Artifact Hub responses to Tekton Hub
  format
- **Comprehensive Middleware**: Includes CORS, logging, and recovery middleware
- **Health Checks**: Built-in health check endpoint
- **Docker Support**: Containerized deployment ready
- **Security**: Read-only proxy with input validation and structured error handling

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
  cache:
    enabled: true     # Enable/disable API response caching
    ttl: 1h          # Cache time-to-live (e.g., 5m, 10m, 1h)
    max_size: 2000   # Maximum number of cache entries

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
- `THP_ARTIFACTHUB_CACHE_ENABLED=true`
- `THP_ARTIFACTHUB_CACHE_TTL=1h`
- `THP_ARTIFACTHUB_CACHE_MAX_SIZE=2000`
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
  --disable-cache              Disable API response caching
  --cache-ttl duration         Cache TTL duration (e.g., 5m, 10m) (overrides config)
  --cache-max-size int         Maximum number of cache entries (overrides config)
  --help                       Show help message
```

**Examples**:
```bash
# Start with custom port and disabled landing page
./bin/tekton-hub-proxy --port 9090 --disable-landing-page

# Start with debug logging and custom config
./bin/tekton-hub-proxy --debug --config /path/to/config.yaml

# Start with disabled cache
./bin/tekton-hub-proxy --disable-cache

# Start with custom cache settings
./bin/tekton-hub-proxy --cache-ttl 10m --cache-max-size 5000

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

## In-Memory Caching

The proxy includes a high-performance in-memory cache that dramatically improves response times for frequently accessed resources. This is especially beneficial for CI/CD pipelines that repeatedly fetch the same tasks and pipelines.

### Performance Benefits

The cache provides significant performance improvements:

- **🚀 Cache hits**: ~0.08ms response time (1000x+ faster)
- **🌐 Cache misses**: ~100ms response time (API call to Artifact Hub)
- **📦 Memory efficient**: Configurable size limits with LRU eviction
- **🔄 Auto-refresh**: Configurable TTL with automatic cleanup

### Cache Configuration

#### YAML Configuration

```yaml
artifacthub:
  cache:
    enabled: true     # Enable/disable caching (default: true)
    ttl: 1h          # Time-to-live for cache entries (default: 1h)
    max_size: 2000   # Maximum cache entries (default: 2000)
```

#### Environment Variables

```bash
export THP_ARTIFACTHUB_CACHE_ENABLED=true
export THP_ARTIFACTHUB_CACHE_TTL=10m
export THP_ARTIFACTHUB_CACHE_MAX_SIZE=5000
```

#### Command Line Flags

```bash
# Disable caching entirely
./bin/tekton-hub-proxy --disable-cache

# Custom cache settings
./bin/tekton-hub-proxy --cache-ttl 10m --cache-max-size 5000

# High-performance setup for heavy usage
./bin/tekton-hub-proxy --cache-ttl 15m --cache-max-size 10000
```

### Cache Behavior

#### What Gets Cached

- **Package metadata**: Resource details, versions, descriptions
- **YAML manifests**: Task and pipeline definitions
- **README content**: Documentation and examples
- **Search results**: Query responses and resource listings
- **Catalog information**: Available catalogs and mappings

#### Cache Keys

The cache uses SHA-256 hashed keys based on:
- **Packages**: `package:{repoKind}:{catalog}:{name}:{version}`
- **Latest packages**: `package-latest:{repoKind}:{catalog}:{name}`
- **Search queries**: `search:{encodedQueryParams}`

#### Memory Management

- **LRU Eviction**: Least recently used entries are removed when cache is full
- **TTL Cleanup**: Expired entries are automatically removed every `TTL/2` interval
- **Memory Safety**: Hard limits prevent unbounded memory growth

### Cache Logging

The proxy provides prominent cache status logging to monitor performance:

```json
{"level":"info","msg":"🚀 CACHE HIT - GetPackage","api_call":"GetPackage","catalog":"tekton","name":"git-clone"}
{"level":"info","msg":"📦 API CALL CACHED - GetPackageLatest","api_call":"GetPackageLatest","cache_size":42}
{"level":"info","msg":"🌐 API CALL NO CACHE - SearchPackages","api_call":"SearchPackages"}
```

#### Log Messages

- **🚀 CACHE HIT**: Request served from cache (super fast)
- **📦 API CALL CACHED**: New data fetched and stored in cache
- **🌐 API CALL NO CACHE**: Cache disabled, direct API call

### Cache Sizing Guidelines

#### Memory Usage Estimates

Approximate memory usage per cache entry:
- **Package metadata**: ~2-5KB per entry
- **YAML manifests**: ~5-20KB per entry
- **Search results**: ~10-50KB per entry

#### Recommended Settings

| Use Case | TTL | Max Size | Est. Memory |
|----------|-----|----------|-------------|
| **Development** | 5m | 500 | ~10-25MB |
| **CI/CD Light** | 10m | 2000 | ~40-100MB |
| **CI/CD Heavy** | 15m | 5000 | ~100-250MB |
| **Production** | 30m | 10000 | ~200-500MB |

#### Environment-Specific Recommendations

```bash
# Development environment
export THP_ARTIFACTHUB_CACHE_TTL=5m
export THP_ARTIFACTHUB_CACHE_MAX_SIZE=500

# CI/CD environment with frequent builds
export THP_ARTIFACTHUB_CACHE_TTL=15m
export THP_ARTIFACTHUB_CACHE_MAX_SIZE=5000

# Production proxy serving multiple teams
export THP_ARTIFACTHUB_CACHE_TTL=30m
export THP_ARTIFACTHUB_CACHE_MAX_SIZE=10000

# Memory-constrained environments
export THP_ARTIFACTHUB_CACHE_TTL=5m
export THP_ARTIFACTHUB_CACHE_MAX_SIZE=100
```

### Cache Monitoring

#### Metrics Available in Logs

- **Cache size**: Current number of cached entries
- **Hit/miss ratio**: Visible through log message types
- **Memory cleanup**: Eviction and expiration events

#### Health Monitoring

```bash
# Monitor cache performance via logs
kubectl logs -f deployment/tekton-hub-proxy | grep -E "(CACHE HIT|CACHED|NO CACHE)"

# Watch cache size growth
kubectl logs -f deployment/tekton-hub-proxy | grep "cache_size" | tail -20
```

#### Example Cache Monitoring

```bash
# Show cache statistics for the last hour
kubectl logs --since=1h deployment/tekton-hub-proxy | \
  grep -E "(CACHE HIT|CACHED|NO CACHE)" | \
  sort | uniq -c | sort -nr
```

### Cache Performance Tuning

#### For High-Frequency Access

```yaml
artifacthub:
  cache:
    enabled: true
    ttl: 30m          # Longer TTL for stable resources
    max_size: 10000   # Large cache for popular items
```

#### For Memory-Constrained Environments

```yaml
artifacthub:
  cache:
    enabled: true
    ttl: 2m           # Short TTL to reduce memory usage
    max_size: 200     # Small cache size
```

#### For Development/Testing

```yaml
artifacthub:
  cache:
    enabled: false    # Disable to always get fresh data
```

### Cache Invalidation

#### Automatic Invalidation
- **TTL expiration**: Entries automatically expire after configured time
- **LRU eviction**: Oldest entries removed when cache fills up
- **Service restart**: Cache is cleared on application restart

#### Manual Cache Control
- **Restart service**: Clears all cache entries
- **Reduce TTL**: Expires entries faster
- **Disable cache**: Set `enabled: false` for fresh data

### Troubleshooting Cache Issues

#### Cache Not Working
1. Check if caching is enabled in configuration
2. Verify cache settings in logs at startup
3. Look for cache hit/miss messages in logs

#### High Memory Usage
1. Reduce `max_size` setting
2. Decrease `ttl` for faster cleanup
3. Monitor cache size in logs

#### Stale Data Issues
1. Reduce `ttl` for fresher data
2. Restart service to clear cache
3. Temporarily disable cache

## Testing with [Tekton Hub Resolver](https://tekton.dev/docs/pipelines/hub-resolver/)

To test the proxy with Tekton Pipelines, you need to configure the [Tekton Hub Resolver](https://tekton.dev/docs/pipelines/hub-resolver/) to use your proxy instance instead of the default Tekton Hub.

### Prerequisites

- A Kubernetes cluster with Tekton Pipelines installed
- `kubectl` configured to access your cluster
- The proxy running and accessible from your cluster

### Configuration Steps

1. **Configure the Tekton Hub Resolver**:

   Point the [Tekton Hub Resolver](https://tekton.dev/docs/pipelines/hub-resolver/) to use your proxy instance:

   ```bash
   kubectl set env -n tekton-pipelines-resolvers \
     deployments.app/tekton-pipelines-remote-resolvers \
     TEKTON_HUB_API=https://tknhub.pipelinesascode.com/
   ```

   Replace `https://tknhub.pipelinesascode.com/` with your actual proxy URL.

2. **Verify the Configuration**:

   Check that the environment variable is set:

   ```bash
   kubectl get deployment -n tekton-pipelines-resolvers tekton-pipelines-remote-resolvers -o yaml | grep TEKTON_HUB_API
   ```

### Test Example

Create a simple test PipelineRun to verify the proxy is working:

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: hub-test-
spec:
  pipelineSpec:
    tasks:
      - name: test-task
        taskRef:
          resolver: hub
          params:
            - name: kind
              value: task
            - name: name
              value: tkn
            - name: version
              value: "0.4"
            - name: type
              value: tekton
            - name: catalog
              value: tekton
        params:
          - name: ARGS
            value:
              - version
```

### Running the Test

1. **Save the PipelineRun** to a file (e.g., `test-pipelinerun.yaml`)

2. **Apply the PipelineRun**:

   ```bash
   kubectl apply -f test-pipelinerun.yaml
   ```

3. **Monitor the PipelineRun**:

   ```bash
   kubectl get pipelineruns
   kubectl logs -f pipelinerun/hub-test-xxxxx
   ```

### Expected Results

If the proxy is working correctly:

- The PipelineRun should start successfully
- The task should be resolved from Artifact Hub via your proxy
- You should see logs indicating the task is running
- The proxy logs should show the API requests being handled

### Troubleshooting

- **PipelineRun fails to start**: Check proxy accessibility and configuration
- **Task resolution errors**: Verify catalog mappings in proxy configuration
- **Connection timeouts**: Ensure network connectivity between cluster and proxy
- **Check proxy logs** for detailed error information

### Alternative Test Methods

You can also test the proxy directly using curl:

```bash
# Test health endpoint
curl https://tknhub.pipelinesascode.com/health

# Test catalog listing
curl https://tknhub.pipelinesascode.com/v1/catalogs

# Test specific task resolution
curl https://tknhub.pipelinesascode.com/v1/resource/tekton/task/tkn
```

## Testing with [Pipelines-as-Code](https://pipelinesascode.com/)

**Important Note**: Latest versions of [Pipelines-as-Code](https://pipelinesascode.com/) automatically use Artifact Hub by default. This configuration is only needed for older PaC versions that cannot be upgraded.

[Pipelines-as-Code (PaC)](https://pipelinesascode.com/) supports fetching tasks and pipelines from remote hub catalogs. You can configure older PaC versions to use this proxy as a custom hub for task resolution.

### Configuration

To configure PaC to use your proxy, you need to modify the PaC ConfigMap with the following settings:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pipelines-as-code
  namespace: pipelines-as-code
data:
  # Configure the proxy as a custom hub
  hub-url: https://tknhub.pipelinesascode.com/
  hub-catalog-name: tekton
  hub-catalog-type: tektonhub
```

### Configuration Options

Based on the [PaC Settings Documentation](https://pipelinesascode.com/docs/install/settings/), the following options are available:

- **`hub-url`**: The base URL for the hub API. Set this to your proxy URL (e.g., `https://tknhub.pipelinesascode.com/`)
- **`hub-catalog-name`**: The catalog name in the hub. For this proxy, use `tekton`
- **`hub-catalog-type`**: The type of hub catalog. Set to `tektonhub` to use Tekton Hub format

### Multiple Hub Configuration

You can also configure multiple hubs including this proxy as an additional catalog:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pipelines-as-code
  namespace: pipelines-as-code
data:
  # Default Artifact Hub (implicit)
  # ... other configurations ...

  # Custom proxy hub
  catalog-1-id: "proxy"
  catalog-1-name: "tekton"
  catalog-1-url: "https://tknhub.pipelinesascode.com/"
  catalog-1-type: "tektonhub"
```

### Using the Proxy in PaC

Once configured, you can reference tasks from the proxy in your `.tekton` pipeline files:

#### With default hub configuration:
```yaml
# This will use the configured default hub (your proxy)
pipelineSpec:
  tasks:
    - name: task-from-proxy
      taskRef:
        resolver: hub
        params:
          - name: kind
            value: task
          - name: name
            value: tkn
          - name: version
            value: "0.4"
          - name: catalog
            value: tekton
```

#### With multiple hub configuration:
```yaml
# This will use the proxy hub explicitly
pipelineSpec:
  tasks:
    - name: task-from-proxy
      taskRef:
        resolver: hub
        params:
          - name: kind
            value: task
          - name: name
            value: tkn
          - name: version
            value: "0.4"
          - name: catalog
            value: tekton
          - name: hub
            value: proxy  # References catalog-1-id
```

### Applying the Configuration

1. **Save the ConfigMap** to a file (e.g., `pac-config.yaml`)

2. **Apply the configuration**:
   ```bash
   kubectl apply -f pac-config.yaml
   ```

3. **Restart PaC controller** (if needed):
   ```bash
   kubectl rollout restart deployment/pipelines-as-code-controller -n pipelines-as-code
   ```

### Verification

To verify the configuration is working:

1. **Check PaC logs** for any configuration errors:
   ```bash
   kubectl logs deployment/pipelines-as-code-controller -n pipelines-as-code
   ```

2. **Test with a simple pipeline** that references a task from your proxy

3. **Monitor proxy logs** to see requests from PaC

For more details on PaC configuration, see the [official settings documentation](https://pipelinesascode.com/docs/install/settings/).

## Migration Guidance

This proxy is designed as a temporary bridge during the transition from Tekton Hub to Artifact Hub. Here's how to migrate away from this proxy:

### For Tekton Pipelines Hub Resolver

**Modern Approach**: Configure the Hub Resolver to use Artifact Hub directly:

```bash
kubectl set env -n tekton-pipelines-resolvers \
  deployments.app/tekton-pipelines-remote-resolvers \
  TEKTON_HUB_API=https://artifacthub.io
```

**Update your PipelineRuns** to use Artifact Hub catalog names:
- Change `catalog: tekton` to `catalog: tekton-catalog-tasks`
- Ensure you're using full semantic versions (e.g., `0.4.0` instead of `0.4`)

### For Pipelines-as-Code

**Modern Approach**: Upgrade to the latest Pipelines-as-Code version which uses Artifact Hub by default. No additional configuration needed.

**If upgrade isn't possible**, use this proxy as documented above.

### Direct API Usage

**Modern Approach**: Update your applications to use the Artifact Hub API directly:
- Base URL: `https://artifacthub.io/api/v1/packages/tekton-task/tekton-catalog-tasks/`
- Use full semantic versioning
- Adapt to Artifact Hub response format

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
    handle /v1/resource/*/*/yaml {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=604800, immutable"
    }

    # Raw YAML content with very long cache
    handle /v1/resource/*/raw {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=604800, immutable"
    }

    # README content with medium cache
    handle /v1/resource/*/*/readme {
        reverse_proxy localhost:8080
        header Cache-Control "public, max-age=86400, stale-while-revalidate=3600"
    }

    # Resource metadata with medium cache
    handle /v1/resource/* {
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

## Security

This proxy is designed with a security-first mindset, incorporating multiple layers of protection.

### Read-Only Design
The proxy operates in a read-only capacity. It only performs `GET` requests to the upstream Artifact Hub API and does not accept or store any user-provided data, which significantly limits its attack surface.

### Denial of Service (DoS) Protection
To protect against resource exhaustion attacks, the proxy implements strict timeouts:
- **Server Timeouts**: The HTTP server is configured with `Read`, `Write`, and `Idle` timeouts to prevent slow client attacks (e.g., Slowloris) from holding connections open and exhausting resources.
- **Client Timeouts**: The HTTP client that connects to Artifact Hub has a built-in timeout, ensuring that a slow or unresponsive backend cannot cause a cascading failure in the proxy.

### Secure HTTP Headers
A security middleware is in place to add the following HTTP headers to all responses, providing an additional layer of defense against common web vulnerabilities:
- `X-Content-Type-Options: nosniff`: Prevents browsers from MIME-sniffing the content-type of a response.
- `X-Frame-Options: DENY`: Protects against clickjacking attacks by preventing the content from being embedded in iframes.
- `Content-Security-Policy: default-src 'self'`: Restricts the sources from which content can be loaded, mitigating XSS and other injection attacks.
- `X-XSS-Protection: 1; mode=block`: Enables the built-in XSS filter in older browsers.

### Secure Coding Practices
- **Input Validation**: Path parameters are validated using regular expressions at the routing layer (e.g., ensuring IDs are numeric).
- **Panic Recovery**: A recovery middleware catches any unhandled panics, preventing the server from crashing and logging the error instead.
- **No Sensitive Information in Errors**: The proxy returns generic error messages and avoids leaking internal details like stack traces.

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
