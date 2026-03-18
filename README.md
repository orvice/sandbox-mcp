# sandbox-mcp

A minimal sandbox MCP server built with [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk).

It exposes a `streamable HTTP` endpoint and provides 3 tools by default:

- `ReadFile`
- `WriteFile`
- `Bash`

It is designed to run well inside Docker, with environment variables for the listen address, sandbox root, and bearer token authentication.

## Local Run

```bash
go run .
```

Default endpoints:

- MCP endpoint: `http://127.0.0.1:8080/mcp`
- health: `http://127.0.0.1:8080/healthz`

Example:

```bash
MCP_ADDR=:8080 \
SANDBOX_ROOT=/workspace \
MCP_AUTH_TOKEN=secret-token \
go run .
```

## Docker Run

```bash
docker build -t sandbox-mcp .

docker run --rm -p 8080:8080 \
  -e MCP_AUTH_TOKEN=secret-token \
  -e SANDBOX_ROOT=/workspace \
  -v "$PWD:/workspace" \
  sandbox-mcp
```

## Docker Compose

Example `compose.yaml`:

```yaml
services:
  sandbox-mcp:
    image: ghcr.io/orvice/sandbox-mcp:main
    ports:
      - "8080:8080"
    environment:
      MCP_ADDR: ":8080"
      MCP_HTTP_PATH: "/mcp"
      MCP_AUTH_TOKEN: "secret-token"
      SANDBOX_ROOT: "/workspace"
      SANDBOX_SHELL: "/bin/bash"
    volumes:
      - /tmp/sandbox-workspace:/workspace
    restart: unless-stopped
```

Start it with:

```bash
docker compose up -d
```

## Environment Variables

- `MCP_ADDR`: HTTP listen address, default `:8080`
- `MCP_HTTP_PATH`: MCP HTTP path, default `/mcp`
- `MCP_AUTH_TOKEN`: when set, requires `Authorization: Bearer <token>`
- `SANDBOX_ROOT`: root directory for file access and command execution, default current directory
- `SANDBOX_SHELL`: shell used by the `Bash` tool, default `bash`
- `MCP_STATELESS`: enable stateless streamable HTTP mode, default `false`
- `MCP_JSON_RESPONSE`: prefer `application/json` responses, default `false`

## MCP Server JSON Example

Example client configuration for a streamable HTTP MCP server with bearer auth:

```json
{
  "mcpServers": {
    "sandbox": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp",
      "headers": {
        "Authorization": "Bearer secret-token"
      }
    }
  }
}
```

If your MCP client uses a different schema, keep the same core values:

- endpoint: `http://127.0.0.1:8080/mcp`
- auth header: `Authorization: Bearer <token>`

## Tools

### `ReadFile`

Request payload:

```json
{
  "path": "relative/or/absolute/path"
}
```

### `WriteFile`

Request payload:

```json
{
  "path": "tmp/hello.txt",
  "content": "hello world",
  "createDirs": true
}
```

### `Bash`

Request payload:

```json
{
  "command": "pwd && ls -la",
  "cwd": ".",
  "timeoutSeconds": 30,
  "env": {
    "FOO": "bar"
  }
}
```

The result includes:

- `cwd`
- `exitCode`
- `stdout`
- `stderr`

## Notes

- All file paths are constrained to `SANDBOX_ROOT` to prevent path escape.
- The `Bash` tool working directory is also constrained to `SANDBOX_ROOT`.
- Command output is truncated to avoid returning excessively large responses.
