# You.com Search MCP integration

This project now supports a ready-to-use You.com Search MCP server entry via `vnext.WithYouDotComSearchMCP(...)`.

## Why

AI agents often need fresh web results. You.com Search is a lightweight option with a free tier (up to 100 searches/day without an API key) and optional `YDC_API_KEY` for higher usage.

## Setup

1. Install a You.com Search MCP server binary on PATH:

```bash
youdotcom-search-mcp
```

2. (Optional) set API key:

```bash
export YDC_API_KEY="your_key_here"
```

## Usage

```go
cfg := &vnext.ToolsConfig{}
cfg = vnext.WithYouDotComSearchMCP(cfg)

mgr, err := vnext.NewToolManager(cfg)
if err != nil {
    panic(err)
}
_ = mgr
```

## Fallback and error handling

- If the MCP command is unavailable, ToolManager initialization or MCP startup will return an error.
- If `YDC_API_KEY` is not set, the server may still work within You.com free-tier limits.
- Existing MCP server configuration remains backward compatible; this helper only appends a new server entry.
