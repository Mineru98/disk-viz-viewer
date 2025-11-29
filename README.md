# Disk Usage Viewer

A web-based disk usage visualization tool written in Go.

## Features

- View disk usage for any directory
- Visualize subdirectory sizes with progress bars
- Navigate through directories by clicking
- Configurable depth for recursive analysis
- Beautiful dark theme UI

## Installation

```bash
go build -o disk-viz-viewer .
```

## Usage

```bash
# Start the server on default port (8080)
./disk-viz-viewer

# Start on a custom port
./disk-viz-viewer -port 3000
```

Then open your browser and navigate to `http://localhost:8080`

## API

### GET /api/analyze

Analyze disk usage for a given path.

**Parameters:**
- `path` (string): The directory path to analyze (default: `/`)
- `depth` (int): How deep to analyze subdirectories (1-5, default: 1)

**Example:**
```bash
curl "http://localhost:8080/api/analyze?path=/home&depth=2"
```

**Response:**
```json
{
  "rootPath": "/home",
  "totalSize": 1234567890,
  "totalStr": "1.14 GB",
  "items": [
    {
      "name": "user",
      "path": "/home/user",
      "size": 1234567890,
      "sizeStr": "1.14 GB",
      "isDir": true,
      "children": []
    }
  ]
}
```

## Project Structure

```
.
├── main.go                    # Application entry point
├── internal/
│   ├── api/
│   │   └── handler.go         # HTTP handlers
│   └── disk/
│       └── usage.go           # Disk usage calculation
└── web/
    └── static/
        └── index.html         # Web UI
```

## License

MIT
