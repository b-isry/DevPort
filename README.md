# DevPort

DevPort is a next-generation Node.js dependency engine that virtualizes `node_modules` using a shared, content-addressable cache. It leverages S3-compatible object storage for fast, deduplicated dependency management across machines and branches.

## Features

- Blazingly fast dependency teleportation for Node.js projects
- S3-backed cache for deduplicated file storage
- Manifest-based tracking of file hashes
- CLI commands for scanning, caching, and rebuilding dependencies
- Configurable via YAML and environment variables

## Getting Started

### Prerequisites

- Go 1.20+
- Access to an S3-compatible storage (e.g., MinIO, AWS S3)

### Configuration

DevPort uses Viper for configuration. It loads settings from (in order):

- `--config` flag (custom YAML file)
- `./.devport.yaml` or `$HOME/.devport/config.yaml`
- `./.devport.secret.yaml` (merged, for secrets)
- Environment variables (e.g., `S3_ENDPOINT`)

Example `.devport.secret.yaml`:

```yaml
s3:
  access_key_id: "YOUR SECRET KEY"
  secret_access_key: "YOUR PASSWORD"

```

### Commands

- `devport scan`  
  Scans `node_modules`, hashes files, uploads unique content to S3, and generates a manifest.

- `devport rebuild`  
  Rebuilds `node_modules` from the manifest and S3 cache.

- `devport test-config`  
  Prints current S3 configuration values for debugging.

#### Flags

- `-c, --config <file>`: Specify a custom config file
- `-v, --verbose`: Enable verbose logging

### Example Usage

```sh
# Scan and cache dependencies
./devport push

# Rebuild node_modules from cache
./devport pull

```

## Development

- Main CLI logic in `cmd/`
- S3 client setup in `cmd/s3_client.go`
- Configuration and root command in `cmd/root.go`

## License

MIT
