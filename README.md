# Thor: High-Performance Log Aggregator

Thor is a high-performance log aggregation system designed for scalability, reliability, and real-time analytics.

Thor is a distributed log aggregation system built in Go, designed to handle high-velocity data streams with a focus on concurrency, throughput, and reliability.  
This repository is currently under active development, focusing on the Phase 1 implementation: building a robust single-node core capable of concurrent log ingestion, structured processing, and segmented storage.


## Why Thor:

Thor is the cooler brother of loki (the famous logging service) , why is it cooler you might say:
- **Bifrost Bridge:** Real-time log transformer
- **Heimdall’s Sight:** Smart schema recognition
....... This list is still **WIP**


## Key Features (Phase 1 - In Progress)

- **Concurrent TCP Ingestion:** High-throughput log reception supporting multiple simultaneous client connections.
- **Worker Pool Processing:** A dedicated pipeline of parser workers to handle log enrichment and sanitization without blocking the ingestion layer.
- **Segmented Storage:** Automatic log rotation and segmented file management to prevent massive file overhead.
- **Graceful Shutdown:** Zero-data-loss shutdown sequence that drains all internal buffers before closing.
- **Real-time Metrics:** Built-in monitoring of throughput, channel health, and connection counts.

## Coming on Phase 2
- Query language for logs
- Rest api to interact with logs (read-only)
## Architecture

```
TCP Client → [ Ingestion Channel ] → [ Parser Workers ] → [ Write Channel ] → [ File Writer ]
```

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Installation & Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/thor.git
   cd thor
   ```

2. Build and start the service:
   ```sh
   docker compose up --build
   ```

### Configuration

The system is configured via `config.yaml`. You can tune performance parameters here before starting the Docker container:

```yaml
server:
  tcp port: 9000   # Ingestion Port
pipeline:
  parser workers: 5   # Number of parser workers spawned (default: 5)
storage:
  log directory: "/var/log/thor"   # logs storage dir inside docker (to be mapped via volumes to your device if needed for backups)
  segment size :5000 # size of single segment file before splitting in kb (default:50000 (50 mb) )
```

## Testing the Pipeline

### Manual Test

Once the container is running, you can send logs directly via Netcat to the exposed port to verify the ingestion flow:
```sh
echo "2024-12-01T10:30:45Z|INFO|api-server|User login successful" | nc localhost 9000
```

## Project Structure

- cmd/server/: Application entry point.
- internal/configs/: Sever Configs
- internal/ingestion/: TCP server logic and connection handling.
- internal/api/: The rest api to query different logs
- internal/pipeline/: Concurrent worker pool and log parsing.
- internal/storage/: Segmented file I/O and rotation logic.
- Dockerfile: Containerization configuration.
- docker compose.yaml: Multi-container orchestration (includes storage volume mapping).

## License
MIT © 2026 Mouloud Hasrane





