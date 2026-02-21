# UniRoute Examples

This directory contains example applications demonstrating how to use UniRoute API.

## Examples

### 1. **Basic Chat Example (Go)**
Simple Go application that sends a chat request to UniRoute.

**File**: `basic-chat-go/main.go`

**Usage**:
```bash
cd examples/basic-chat-go
go run main.go
```

### 2. **Basic Chat Example (Python)**
Simple Python application that sends a chat request to UniRoute.

**File**: `basic-chat-python/main.py`

**Usage**:
```bash
cd examples/basic-chat-python
python main.py
```

### 3. **Basic Chat Example (JavaScript/Node.js)**
Simple Node.js application that sends a chat request to UniRoute.

**File**: `basic-chat-nodejs/index.js`

**Usage**:
```bash
cd examples/basic-chat-nodejs
npm install
node index.js
```

### 4. **Tunnel Example**
Example showing how to use UniRoute CLI to tunnel a local service.

**File**: `tunnel-example/README.md`

**Usage**:
```bash
cd examples/tunnel-example
# Follow instructions in README.md
```

## Prerequisites

1. **UniRoute Server Running**
   - Local: `http://localhost:8084`
   - Or production: `https://app.uniroute.co`

2. **API Key**
   - Create via CLI: `uniroute keys create`
   - Or via web UI: Dashboard → API Keys

3. **Environment Variables**
   - `UNIROUTE_API_URL`: API server URL (when you don't use `--server`/`--local`/`--live`: env → saved config → default hosted)
   - `UNIROUTE_API_KEY`: Your API key

## Quick Start

1. **Get your API key**:
   ```bash
   uniroute keys create --name "Example App"
   ```

2. **Set environment variable**:
   ```bash
   export UNIROUTE_API_KEY="ur_your_api_key_here"
   export UNIROUTE_API_URL="http://localhost:8084"  # Optional, for local dev
   ```

3. **Run an example**:
   ```bash
   cd examples/basic-chat-go
   go run main.go
   ```

## API Documentation

- **Swagger UI**: http://localhost:8084/swagger (when server is running)
- **OpenAPI Spec**: http://localhost:8084/swagger.json

## Support

For issues or questions:
- GitHub Issues: https://github.com/Kizsoft-Solution-Limited/uniroute/issues
- Documentation: https://github.com/Kizsoft-Solution-Limited/uniroute/tree/main/docs
