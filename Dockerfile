FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /out/sandbox-mcp .

FROM alpine:3

RUN apk add --no-cache bash ca-certificates git go python3

RUN adduser -D -u 10001 -s /bin/bash appuser \
	&& mkdir -p /workspace \
	&& chown -R appuser:appuser /workspace /home/appuser

WORKDIR /workspace

COPY --from=builder /out/sandbox-mcp /usr/local/bin/sandbox-mcp

ENV MCP_ADDR=:8080
ENV MCP_HTTP_PATH=/mcp
ENV SANDBOX_ROOT=/workspace
ENV SANDBOX_SHELL=/bin/bash

EXPOSE 8080

USER appuser

ENTRYPOINT ["/usr/local/bin/sandbox-mcp"]
