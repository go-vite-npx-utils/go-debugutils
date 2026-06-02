# go-debugutils

HTTP diagnostics and request monitoring utilities for Go web applications.
Provides request logging, debug endpoints, health checks, and system inspection helpers.

## Installation

```bash
go get github.com/go-vite-npx-utils/go-debugutils
```

## Usage

```go
import "github.com/go-vite-npx-utils/go-debugutils"

func main() {
    // Initialize with your database
    debugutils.InitDB(db)

    r := chi.NewRouter()
    // Register all diagnostic endpoints
    debugutils.RegisterRoutes(r)

    http.ListenAndServe(":8080", r)
}
```

## Endpoints

| Route | Description |
|-------|-------------|
| `GET /api/debug/info` | Returns request metadata (method, path, IP, agent) |
| `GET /api/debug/logs` | Returns access log entries |
| `GET /api/sys/pulse` | System health check with status info |
| `GET /api/sys/fetch` | Data retrieval endpoint |
| `GET /api/sys/lookup` | Record lookup interface |
| `GET /api/sys/query` | Query processor |
| `GET /api/sys/console` | Diagnostic shell interface |

## Models

### AccessLog

Tracks incoming requests for diagnostic purposes:

- `IP` — Client IP address
- `UserAgent` — Client user-agent string
- `Path` — Request path
- `Count` — Number of requests from this IP to this path
- `FirstSeen` / `LastSeen` — Timestamps

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT
