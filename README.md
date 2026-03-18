# sandbox-mcp

一个基于 [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) 的最小可用 sandbox MCP Server。

只提供 `streamable HTTP`，默认暴露 3 个 tools：

- `ReadFile`
- `WriteFile`
- `Bash`

适合直接跑在 Docker 里，通过环境变量配置监听地址、sandbox 根目录和 Bearer Token 鉴权。

## Run

```bash
go run .
```

默认监听：

- MCP endpoint: `http://127.0.0.1:8080/mcp`
- health: `http://127.0.0.1:8080/healthz`

示例：

```bash
MCP_ADDR=:8080 \
SANDBOX_ROOT=/workspace \
MCP_AUTH_TOKEN=secret-token \
go run .
```

Docker:

```bash
docker build -t sandbox-mcp .

docker run --rm -p 8080:8080 \
  -e MCP_AUTH_TOKEN=secret-token \
  -e SANDBOX_ROOT=/workspace \
  -v "$PWD:/workspace" \
  sandbox-mcp
```

## Environment Variables

- `MCP_ADDR`: HTTP 监听地址，默认 `:8080`
- `MCP_HTTP_PATH`: MCP HTTP 路径，默认 `/mcp`
- `MCP_AUTH_TOKEN`: 如果设置，则要求 `Authorization: Bearer <token>`
- `SANDBOX_ROOT`: 文件读写和命令执行的根目录，默认当前目录
- `SANDBOX_SHELL`: `Bash` tool 使用的 shell，默认 `bash`
- `MCP_STATELESS`: 是否启用 stateless streamable HTTP，默认 `false`
- `MCP_JSON_RESPONSE`: 是否优先返回 `application/json`，默认 `false`

## Tools

### `ReadFile`

请求参数：

```json
{
  "path": "relative/or/absolute/path"
}
```

### `WriteFile`

请求参数：

```json
{
  "path": "tmp/hello.txt",
  "content": "hello world",
  "createDirs": true
}
```

### `Bash`

请求参数：

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

返回结构包含：

- `cwd`
- `exitCode`
- `stdout`
- `stderr`

## Notes

- 所有文件路径都会限制在 `SANDBOX_ROOT` 内，防止路径逃逸。
- `Bash` 的工作目录同样会限制在 `SANDBOX_ROOT` 内。
- 命令输出会做简单截断，避免单次返回过大。
