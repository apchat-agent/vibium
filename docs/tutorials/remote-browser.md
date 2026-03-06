# Remote Browser Control

Run Chrome on one machine, control it from another.

---

## Server (the machine with the browser)

Install vibium (this downloads Chrome + chromedriver automatically):

```bash
npm install -g vibium
```

Find the chromedriver path and start it:

```bash
vibium paths
# Chromedriver: /Users/you/.cache/vibium/.../chromedriver

$(vibium paths | grep Chromedriver | cut -d' ' -f2) --port=9515 --allowed-ips=""
```

---

## Client (your dev machine)

Install vibium locally — this gives you both the CLI (`npx vibium`) and the JS library:

```bash
npm install vibium
```

Or install globally for a bare `vibium` command:

```bash
npm install -g vibium
```

### CLI

```bash
# One-liner with env var (simplest)
export VIBIUM_CONNECT_URL=ws://your-server:9515/session
vibium go https://example.com
vibium title        # "Example Domain"
vibium text h1      # "Example Domain"
```

```bash
# Or use the connect command
vibium connect ws://your-server:9515/session
vibium go https://example.com
vibium title
vibium disconnect
```

### MCP Server

The MCP server reads the same env vars, so AI agents can use a remote browser:

```bash
VIBIUM_CONNECT_URL=ws://your-server:9515/session vibium mcp
```

Or in your Claude Desktop / Claude Code config:

```json
{
  "mcpServers": {
    "vibium": {
      "command": "vibium",
      "args": ["mcp"],
      "env": {
        "VIBIUM_CONNECT_URL": "ws://your-server:9515/session"
      }
    }
  }
}
```

### JavaScript

```javascript
const { browser } = require('vibium/sync')

const bro = browser.connect('ws://your-server:9515/session')
const page = bro.page()

page.go('https://example.com')
console.log(page.title())        // "Example Domain"
console.log(page.find('h1').text())  // "Example Domain"

bro.close()
```

### Python

```bash
pip install vibium
```

```python
from vibium.sync_api import browser

bro = browser.connect("ws://your-server:9515/session")
page = bro.page()

page.go("https://example.com")
print(page.title())          # "Example Domain"
print(page.find("h1").text())    # "Example Domain"

bro.close()
```

---

## With Authentication

If your endpoint requires auth headers (e.g. a cloud browser provider):

**CLI / MCP** — set `VIBIUM_CONNECT_API_KEY` to send a `Bearer` token:

```bash
export VIBIUM_CONNECT_URL=wss://cloud.example.com/session
export VIBIUM_CONNECT_API_KEY=my-token
vibium go https://example.com
```

Or pass headers explicitly with the daemon:

```bash
vibium daemon start --connect wss://cloud.example.com/session \
  --connect-header "Authorization: Bearer my-token"
```

**JavaScript:**

```javascript
const bro = browser.connect('wss://cloud.example.com/bidi', {
  headers: { 'Authorization': 'Bearer my-token' }
})
```

**Python:**

```python
bro = browser.connect("wss://cloud.example.com/bidi", headers={
    "Authorization": "Bearer my-token",
})
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `VIBIUM_CONNECT_URL` | Remote BiDi WebSocket endpoint (e.g. `ws://host:9515/session`) |
| `VIBIUM_CONNECT_API_KEY` | Sent as `Authorization: Bearer <key>` |

These work everywhere — CLI commands, daemon auto-start, and the MCP server.

---

## How It Works

```
┌────────── Your Machine ──────────┐              ┌──── Remote Machine ─────┐
│                                  │              │                         │
│  ┌──────────┐    ┌──────────┐    │  WebSocket   │    ┌─────────────┐      │
│  │ your code│◄──►│  vibium  │◄───┼──────────────┼───►│ chromedriver│      │
│  └──────────┘    └──────────┘    │              │    └──────┬──────┘      │
│                                  │              │           │             │
│                                  │              │    ┌──────▼──────┐      │
│                                  │              │    │   Chrome    │      │
│                                  │              │    └─────────────┘      │
└──────────────────────────────────┘              └─────────────────────────┘
```

Your code talks to a local vibium process, which proxies to the remote chromedriver over WebSocket. The transport between your code and vibium depends on the interface: IPC for CLI, stdin/stdout pipes for JS/Python clients.

All vibium features (auto-wait, screenshots, tracing) work over remote connections.
