# Blitz Technical Documentation

A comprehensive breakdown of the Blitz codebase, architecture, and implementation details.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Backend (main.go)](#backend-maingo)
3. [Frontend (remote.html)](#frontend-remotehtml)
4. [Communication Protocol](#communication-protocol)
5. [Security Model](#security-model)
6. [Concurrency & Thread Safety](#concurrency--thread-safety)
7. [Music Integration](#music-integration)
8. [Error Handling](#error-handling)
9. [Flow Diagrams](#flow-diagrams)

---

## Architecture Overview

Blitz follows a **client-server architecture** using WebSocket for bidirectional, real-time communication.

```
┌──────────────────┐         WebSocket          ┌──────────────────┐
│                  │◄──────────────────────────►│                  │
│  Web Client      │    JSON Messages (Port     │  Go Server       │
│  (remote.html)   │         8765)              │  (main.go)       │
│                  │◄──────────────────────────►│                  │
└──────────────────┘                            └────────┬─────────┘
                                                         │
                                                         │ Executes
                                                         │
                                                         ▼
                                                ┌─────────────────┐
                                                │ System Commands │
                                                │   & playerctl   │
                                                └─────────────────┘
```

### Key Components

- **WebSocket Server**: Handles persistent connections from clients
- **Command Handler**: Validates and executes allowed commands
- **Music Poller**: Periodically fetches music player status
- **Web Interface**: Provides UI for connection and command execution

---

## Backend (main.go)

### 1. Imports and Dependencies

```go
import (
    "encoding/json"      // JSON marshaling/unmarshaling
    "log"               // Logging
    "net/http"          // HTTP server
    "os/exec"           // Command execution
    "strings"           // String manipulation
    "time"              // Timers and intervals

    "github.com/gorilla/websocket"  // WebSocket implementation
)
```

**Dependencies:**

- `github.com/gorilla/websocket`: Industry-standard WebSocket library for Go

### 2. Global Configuration

#### ALLOWED_COMMANDS Map

```go
var ALLOWED_COMMANDS = map[string][]string{
    "update":       {"sudo", "pacman", "-Syu"},
    "list_home":    {"ls", "-l", "/home/swap/"},
    "status":       {"git", "status"},
    "open_firefox": {"firefox", "--new-window"},
    "open_edge":    {"microsoft-edge-beta"},
    "open_vscode":  {"code-insiders"},
    "open_postman": {"postman"},
}
```

**Purpose:** Security allowlist that maps command names to executable commands.

**Structure:**

- **Key**: Command identifier (sent from client)
- **Value**: Array of strings where:
  - Index 0: Executable name
  - Index 1+: Arguments

**Security:** This is the **primary security mechanism**. Only pre-approved commands can be executed.

#### WebSocket Upgrader

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true  // Allow all origins
    },
}
```

**Purpose:** Upgrades HTTP connections to WebSocket connections.

**CheckOrigin:** Currently allows all origins (`return true`). In production, this should validate the origin:

```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "http://trusted-domain.com"
}
```

### 3. Data Structures

#### ClientMessage

```go
type ClientMessage struct {
    Command string `json:"command"`
}
```

**Incoming message format** from clients.

**Fields:**

- `Command`: The command name to execute (must match a key in `ALLOWED_COMMANDS`)

**Example JSON:**

```json
{ "command": "open_firefox" }
```

#### ServerResponse

```go
type ServerResponse struct {
    Status  string `json:"status"`           // Required: "success", "error", or "player"
    Command string `json:"command,omitempty"` // Optional: echoes command name
    Output  string `json:"output,omitempty"`  // Optional: command output or message
    Message string `json:"message,omitempty"` // Optional: error messages
    Artwork string `json:"artwork,omitempty"` // Optional: album artwork URL
}
```

**Outgoing message format** to clients.

**Status Values:**

- `"success"`: Command executed successfully
- `"error"`: Command failed or not allowed
- `"player"`: Music player status update

**omitempty tag:** Fields are omitted from JSON if empty, reducing message size.

**Example JSON:**

```json
{
  "status": "player",
  "output": "Artist Name - Song Title — Playing",
  "artwork": "file:///path/to/artwork.jpg"
}
```

### 4. Main WebSocket Handler (wsHandler)

This is the core function handling each client connection.

#### Connection Upgrade

```go
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    log.Println("Failed to upgrade connection:", err)
    return
}
defer conn.Close()
log.Printf("Client connected: %s", conn.RemoteAddr())
```

**Process:**

1. Upgrades HTTP request to WebSocket
2. Logs connection establishment
3. Ensures connection closure on function exit (`defer`)

#### Message Channel (Thread-Safe Writing)

```go
messages := make(chan ServerResponse)

go func() {
    for resp := range messages {
        if err := conn.WriteJSON(resp); err != nil {
            log.Println("Failed to write JSON response:", err)
            return
        }
    }
}()
```

**Purpose:** Prevents concurrent writes to the WebSocket connection.

**Why?** WebSocket connections are **not thread-safe** for concurrent writes. Multiple goroutines writing simultaneously can cause race conditions.

**Solution:**

- All messages are sent to a **channel**
- A single **writer goroutine** reads from the channel and writes to WebSocket
- This ensures **serialized writes**

#### Music Poller Goroutine

```go
quitPlayerPoll := make(chan struct{})
go func() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            info, artwork, _ := getPlayerInfo()
            messages <- ServerResponse{Status: "player", Output: info, Artwork: artwork}
        case <-quitPlayerPoll:
            return
        }
    }
}()
```

**Purpose:** Automatically sends music player updates every 3 seconds.

**Components:**

- `ticker`: Fires every 3 seconds
- `select` statement: Waits for either:
  - Ticker event → fetch and send player info
  - Quit signal → exit goroutine
- `defer ticker.Stop()`: Cleanup on exit

**Lifecycle:** Runs for the duration of the WebSocket connection.

#### Message Reading Loop

```go
for {
    messageType, p, err := conn.ReadMessage()
    if err != nil {
        log.Println("Client disconnected:", err)
        break
    }

    if messageType == websocket.TextMessage {
        // Process message
    }
}
```

**Process:**

1. Blocks waiting for incoming messages
2. Breaks loop on error (client disconnect)
3. Only processes text messages (JSON)

#### JSON Parsing

```go
var msg ClientMessage
if err := json.Unmarshal(p, &msg); err != nil {
    log.Println("Error parsing JSON:", err)
    response = ServerResponse{Status: "error", Message: "Invalid JSON format"}
    messages <- response
    continue
}
```

**Error Handling:** Invalid JSON returns an error response to client.

#### Special Command: player_info

```go
if msg.Command == "player_info" {
    info, artwork, _ := getPlayerInfo()
    messages <- ServerResponse{Status: "player", Output: info, Artwork: artwork}
    continue
}
```

**Purpose:** Allows client to request immediate player info (instead of waiting for next 3-second poll).

#### Command Validation and Execution

```go
if commandToRun, ok := ALLOWED_COMMANDS[msg.Command]; ok {
    log.Printf("Running command: %v", commandToRun)

    cmd := exec.Command(commandToRun[0], commandToRun[1:]...)
    err := cmd.Start()

    if err != nil {
        response = ServerResponse{
            Status:  "error",
            Message: err.Error(),
        }
    } else {
        response = ServerResponse{
            Status:  "success",
            Command: msg.Command,
            Output:  "Command launched successfully.",
        }
    }
} else {
    response = ServerResponse{
        Status:  "error",
        Message: "Command '" + msg.Command + "' is not in the allowed list.",
    }
}
```

**Security Check:**

- Uses map lookup to verify command is in allowlist
- Rejects unknown commands immediately

**Execution:**

- Uses `cmd.Start()` instead of `cmd.Run()`
- **Non-blocking**: Launches command in background
- Useful for GUI applications that should stay open

**Why Start() vs Run()?**

- `Start()`: Launches process and returns immediately
- `Run()`: Waits for process to complete before returning
- For apps like Firefox, we want `Start()` so server doesn't block

#### Connection Cleanup

```go
close(quitPlayerPoll)
close(messages)
```

**Order matters:**

1. Signal music poller to quit
2. Close message channel (stops writer goroutine)

### 5. Music Player Integration (getPlayerInfo)

```go
func getPlayerInfo() (string, string, error) {
    // Fetch track metadata
    metaCmd := exec.Command("playerctl", "metadata", "--format", "{{artist}} - {{title}}")
    outMeta, errMeta := metaCmd.Output()
    meta := strings.TrimSpace(string(outMeta))

    // Fetch playback status
    statusCmd := exec.Command("playerctl", "status")
    outStatus, errStatus := statusCmd.Output()
    status := strings.TrimSpace(string(outStatus))

    // Fetch artwork URL
    artworkCmd := exec.Command("playerctl", "metadata", "mpris:artUrl")
    outArtwork, _ := artworkCmd.Output()
    artwork := strings.TrimSpace(string(outArtwork))

    // Error handling
    if errMeta != nil && errStatus != nil {
        return "playerctl not available or no player running", "", errMeta
    }

    // Build info string
    if meta == "" && status == "" {
        return "No player running", "", nil
    }

    info := meta
    if info == "" {
        info = "(unknown track)"
    }
    if status != "" {
        info = info + " — " + status
    }

    return info, artwork, nil
}
```

**Return Values:**

1. **info** (string): Human-readable track info (e.g., "Artist - Title — Playing")
2. **artwork** (string): File path or URL to album art
3. **error**: Non-nil if playerctl unavailable

**playerctl Commands:**

- `playerctl metadata --format "..."`: Gets track metadata with custom format
- `playerctl status`: Returns "Playing", "Paused", or "Stopped"
- `playerctl metadata mpris:artUrl`: Gets artwork file path/URL

**MPRIS:** Media Player Remote Interfacing Specification - Linux standard for media player control.

**Graceful Degradation:**

- If playerctl fails → returns friendly message
- If no metadata → shows "(unknown track)"
- Artwork failure is ignored (doesn't error out)

### 6. Main Function

```go
func main() {
    http.HandleFunc("/ws", wsHandler)

    log.Println("⚡ Blitz server starting at ws://0.0.0.0:8765/ws")
    err := http.ListenAndServe("0.0.0.0:8765", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
```

**Setup:**

- Registers `/ws` endpoint with `wsHandler`
- Listens on `0.0.0.0:8765` (all network interfaces)
- Blocks until server crashes

**Network Binding:**

- `0.0.0.0`: Accepts connections from any network interface
- Alternative `127.0.0.1`: Only local connections
- Port `8765`: Arbitrary choice, can be changed

---

## Frontend (remote.html)

### 1. HTML Structure

```html
<div class="container">
  <h1>⚡ Blitz Remote</h1>

  <!-- Connection controls -->
  <div class="connection-area">
    <input type="text" id="ipAddress" placeholder="Enter PC IP..." />
    <button id="connectButton">Connect</button>
  </div>

  <!-- Connection status -->
  <p id="status" class="disconnected">Status: Not Connected</p>

  <!-- Music player display -->
  <div class="player-info">
    <h3>🎵 Now Playing</h3>
    <div class="player-content">
      <img id="albumArt" class="album-art" alt="Album artwork" />
      <div class="track" id="playerInfo">No music info yet...</div>
    </div>
  </div>

  <!-- Command buttons -->
  <div id="remote-grid">
    <button class="remote-btn" onclick="sendCommand('update')">
      🚀 System Update
    </button>
    <!-- ... more buttons ... -->
  </div>

  <!-- Output log -->
  <h3>Output Log:</h3>
  <pre id="output"></pre>
</div>
```

**Key Elements:**

- `#ipAddress`: Input for server IP
- `#connectButton`: Initiates WebSocket connection
- `#status`: Shows connection state
- `#albumArt`: Album artwork image
- `#playerInfo`: Track information text
- `#remote-grid`: Command buttons
- `#output`: Command execution results

### 2. CSS Styling

#### CSS Variables (Design Tokens)

```css
:root {
  --bg-dark: #1a1a2e; /* Main background */
  --bg-light: #16213e; /* Container background */
  --accent-primary: #0f3460; /* Buttons, borders */
  --accent-secondary: #e94560; /* Hover states */
  --text-light: #e0e0e0; /* Primary text */
  --text-dark: #a0a0a0; /* Secondary text */
  --status-connected: #2ecc71; /* Green for connected */
  --status-disconnected: #e74c3c; /* Red for disconnected */
  --font-main: -apple-system, ...; /* System font stack */
  --font-mono: "SFMono-Regular", ...; /* Monospace for logs */
}
```

**Benefits:**

- Centralized theming
- Easy color scheme changes
- Consistent design language

#### Layout - Centered Container

```css
body {
  display: grid;
  place-items: center; /* Centers both horizontally and vertically */
  min-height: 100vh; /* Full viewport height */
}
```

**Modern CSS Grid:** Single line for perfect centering.

#### Button Grid Layout

```css
#remote-grid {
  display: grid;
  grid-template-columns: 1fr 1fr; /* 2 equal columns */
  gap: 15px;
}
```

**Responsive 2-column grid** for command buttons.

#### Album Art Visibility Toggle

```css
.album-art {
  display: none; /* Hidden by default */
}

.album-art.visible {
  display: block; /* Shown when artwork available */
}
```

**JavaScript controls visibility** by adding/removing `visible` class.

#### Interactive Button States

```css
.remote-btn:hover {
  background-color: var(--accent-secondary);
  transform: translateY(-2px); /* Lift effect */
  box-shadow: 0 5px 15px rgba(233, 69, 96, 0.2); /* Glow */
}

.remote-btn:active {
  transform: scale(0.98); /* Press-down effect */
}
```

**Smooth animations** enhance user feedback.

### 3. JavaScript Logic

#### Global State

```javascript
let ws = null; // WebSocket connection instance
const statusEl = document.getElementById("status");
const outputEl = document.getElementById("output");
const ipInput = document.getElementById("ipAddress");
const connectBtn = document.getElementById("connectButton");
```

**Single source of truth** for WebSocket connection.

#### Connection Function

```javascript
function connect() {
  const ip = ipInput.value.trim();
  const port = 8765;

  // Validation
  if (!ip) {
    outputEl.textContent = "Please enter an IP address.";
    return;
  }

  // Close existing connection
  if (ws) {
    ws.close();
  }

  // Update UI state
  statusEl.textContent = "Connecting...";
  statusEl.className = "";
  connectBtn.disabled = true;
  ipInput.disabled = true;

  // Create WebSocket connection
  ws = new WebSocket(`ws://${ip}:${port}/ws`);

  // Attach event handlers
  ws.onopen = handleOpen;
  ws.onmessage = handleMessage;
  ws.onclose = handleClose;
  ws.onerror = handleError;
}
```

**Process:**

1. Validate IP input
2. Close existing connection (prevent duplicates)
3. Disable UI (prevent multiple connection attempts)
4. Create new WebSocket
5. Attach event handlers

#### WebSocket Event: onopen

```javascript
ws.onopen = () => {
  console.log("Connected to WebSocket server");
  statusEl.textContent = "Status: Connected";
  statusEl.className = "connected"; // Green styling
};
```

**Triggered:** When WebSocket connection successfully established.

#### WebSocket Event: onmessage

```javascript
ws.onmessage = (event) => {
  console.log("Message from server:", event.data);
  let data = JSON.parse(event.data);

  if (data.status === "player") {
    // Update music player display
    const playerInfoEl = document.getElementById("playerInfo");
    const albumArtEl = document.getElementById("albumArt");

    playerInfoEl.textContent = data.output || "No music playing";

    // Show/hide album artwork
    if (data.artwork && data.artwork !== "") {
      albumArtEl.src = data.artwork;
      albumArtEl.classList.add("visible");
    } else {
      albumArtEl.classList.remove("visible");
    }
  } else if (data.status === "success") {
    outputEl.textContent = data.output;
  } else {
    outputEl.textContent = `Error: ${data.message}\n\n${data.output || ""}`;
  }
};
```

**Message Handling:**

- **Player updates**: Update music display and artwork
- **Success**: Show command output
- **Error**: Display error message

**Artwork Logic:**

- If artwork URL exists → set `src` and show image
- If no artwork → hide image (cleaner UI)

#### WebSocket Event: onclose

```javascript
ws.onclose = () => {
  console.log("Disconnected from WebSocket server");
  statusEl.textContent = "Status: Disconnected";
  statusEl.className = "disconnected"; // Red styling
  connectBtn.disabled = false;
  ipInput.disabled = false;
  ws = null; // Clear reference
};
```

**Cleanup:**

- Update UI to reflect disconnection
- Re-enable controls for reconnection
- Clear WebSocket reference

#### WebSocket Event: onerror

```javascript
ws.onerror = (error) => {
  console.error("WebSocket Error:", error);
  statusEl.textContent = "Status: Error (Check IP/Firewall)";
  statusEl.className = "disconnected";
  connectBtn.disabled = false;
  ipInput.disabled = false;
  ws = null;
};
```

**Common causes:**

- Wrong IP address
- Server not running
- Firewall blocking connection
- Network issues

#### Command Sending

```javascript
function sendCommand(commandName) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    const message = {
      command: commandName,
    };
    ws.send(JSON.stringify(message));
    outputEl.textContent = `Sent command: ${commandName}...`;
  } else {
    outputEl.textContent = "Not connected. Please connect first.";
  }
}
```

**Safety checks:**

- Verify WebSocket exists
- Verify connection is open (`readyState === WebSocket.OPEN`)

**WebSocket States:**

- `0` (CONNECTING): Connection not yet established
- `1` (OPEN): Ready to send/receive
- `2` (CLOSING): Connection is closing
- `3` (CLOSED): Connection closed

#### Event Listeners

```javascript
connectBtn.addEventListener("click", connect);

ipInput.addEventListener("keypress", (e) => {
  if (e.key === "Enter") {
    connect(); // Allow Enter key to connect
  }
});
```

**UX Enhancement:** Enter key triggers connection (keyboard-friendly).

---

## Communication Protocol

### Message Flow

#### 1. Client → Server: Command Request

```json
{
  "command": "open_firefox"
}
```

#### 2. Server → Client: Success Response

```json
{
  "status": "success",
  "command": "open_firefox",
  "output": "Command launched successfully."
}
```

#### 3. Server → Client: Error Response

```json
{
  "status": "error",
  "message": "Command 'malicious' is not in the allowed list."
}
```

#### 4. Server → Client: Player Update (Automatic, every 3s)

```json
{
  "status": "player",
  "output": "Pink Floyd - Time — Playing",
  "artwork": "file:///home/user/.cache/spotify/artwork.jpg"
}
```

### WebSocket Lifecycle

```
┌────────────┐                                  ┌────────────┐
│   Client   │                                  │   Server   │
└─────┬──────┘                                  └──────┬─────┘
      │                                                │
      │  HTTP GET /ws (Upgrade: websocket)             │
      │───────────────────────────────────────────────►│
      │                                                │
      │  HTTP 101 Switching Protocols                  │
      │◄───────────────────────────────────────────────│
      │                                                │
      │         WebSocket Connection Established       │
      │◄──────────────────────────────────────────────►│
      │                                                │
      │          Player Update (every 3s)              │
      │◄───────────────────────────────────────────────│
      │                                                │
      │  {"command": "open_firefox"}                   │
      │───────────────────────────────────────────────►│
      │                                                │
      │  {"status": "success", ...}                    │
      │◄───────────────────────────────────────────────│
      │                                                │
      │          Player Update (every 3s)              │
      │◄───────────────────────────────────────────────│
      │                                                │
      │  Close Frame                                   │
      │───────────────────────────────────────────────►│
      │                                                │
      │  Close Frame                                   │
      │◄───────────────────────────────────────────────│
      │                                                │
```

---

## Security Model

### 1. Command Allowlist

**Mechanism:** Map-based allowlist (`ALLOWED_COMMANDS`)

**Protection Against:**

- Arbitrary command execution
- Command injection
- Privilege escalation (if configured properly)

**Limitations:**

- Still vulnerable if allowed commands have exploits
- No rate limiting
- No authentication

### 2. Missing Security Features

⚠️ **Current vulnerabilities:**

1. **No Authentication**

   - Anyone on network can connect
   - No user verification

2. **No Authorization**

   - All connected clients have same permissions

3. **No Rate Limiting**

   - Vulnerable to command spam
   - Could DOS system resources

4. **Unencrypted WebSocket**

   - Uses `ws://` instead of `wss://`
   - Traffic visible on network

5. **Open CORS**
   - `CheckOrigin` returns `true` for all origins
   - Any website can connect

### 3. Security Improvements (Recommendations)

#### Add Authentication

```go
type ClientMessage struct {
    Command string `json:"command"`
    Token   string `json:"token"`  // Add authentication token
}

// Validate token before executing commands
if msg.Token != os.Getenv("AUTH_TOKEN") {
    // Reject
}
```

#### Implement Rate Limiting

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Every(time.Second), 5)  // 5 commands/sec

if !limiter.Allow() {
    // Reject request
}
```

#### Use TLS (wss://)

```go
err := http.ListenAndServeTLS("0.0.0.0:8765", "cert.pem", "key.pem", nil)
```

#### Restrict CORS

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{"http://trusted.com"}
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }
        return false
    },
}
```

---

## Concurrency & Thread Safety

### Goroutines per Connection

Each WebSocket connection spawns **2 goroutines**:

1. **Writer Goroutine**

   - Reads from `messages` channel
   - Writes to WebSocket connection
   - Ensures serialized writes

2. **Music Poller Goroutine**
   - Runs every 3 seconds
   - Fetches player info
   - Sends to `messages` channel

Plus the **main goroutine** that:

- Reads incoming messages
- Processes commands
- Sends responses to `messages` channel

### Why Channels?

**Problem:** WebSocket is not safe for concurrent writes.

**Solution:** Channel-based message queue.

```go
// Multiple goroutines can safely send here
messages <- ServerResponse{...}

// Single goroutine reads and writes
for resp := range messages {
    conn.WriteJSON(resp)  // Only one writer
}
```

### Cleanup Process

```
Connection ends → Main loop exits
                 ↓
          close(quitPlayerPoll)  → Music poller goroutine exits
                 ↓
          close(messages)        → Writer goroutine exits (channel closed)
                 ↓
          defer conn.Close()     → WebSocket connection closed
```

**Order is critical** to prevent:

- Deadlocks
- Resource leaks
- Panic from closed channels

---

## Music Integration

### playerctl Overview

**playerctl** is a command-line utility for controlling MPRIS-compatible media players.

**Installation:**

```bash
# Arch Linux
sudo pacman -S playerctl

# Ubuntu/Debian
sudo apt install playerctl
```

### Commands Used

1. **Get Track Metadata**

   ```bash
   playerctl metadata --format "{{artist}} - {{title}}"
   ```

   Example output: `Pink Floyd - Time`

2. **Get Playback Status**

   ```bash
   playerctl status
   ```

   Output: `Playing`, `Paused`, or `Stopped`

3. **Get Album Artwork**
   ```bash
   playerctl metadata mpris:artUrl
   ```
   Example output: `file:///home/user/.cache/artwork.jpg`

### MPRIS Protocol

**MPRIS** (Media Player Remote Interfacing Specification):

- D-Bus interface specification
- Standard for media player control on Linux
- Supported by: Spotify, VLC, Firefox, Chrome, etc.

### Update Frequency

**Current:** Every 3 seconds

**Why not faster?**

- Reduces CPU usage
- playerctl execution overhead
- Music changes don't happen that frequently

**Configurable:**

```go
ticker := time.NewTicker(3 * time.Second)  // Change duration here
```

---

## Error Handling

### Backend Error Handling

1. **Connection Upgrade Failure**

   ```go
   conn, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
       log.Println("Failed to upgrade connection:", err)
       return  // Early exit
   }
   ```

2. **JSON Parsing Error**

   ```go
   if err := json.Unmarshal(p, &msg); err != nil {
       response = ServerResponse{Status: "error", Message: "Invalid JSON format"}
       messages <- response
       continue  // Skip this message, continue listening
   }
   ```

3. **Command Execution Error**

   ```go
   err := cmd.Start()
   if err != nil {
       response = ServerResponse{
           Status:  "error",
           Message: err.Error(),  // Include system error message
       }
   }
   ```

4. **playerctl Unavailable**
   ```go
   if errMeta != nil && errStatus != nil {
       return "playerctl not available or no player running", "", errMeta
   }
   ```
   **Graceful degradation**: Returns friendly message instead of crashing.

### Frontend Error Handling

1. **Missing IP Address**

   ```javascript
   if (!ip) {
     outputEl.textContent = "Please enter an IP address.";
     return;
   }
   ```

2. **Connection Errors**

   ```javascript
   ws.onerror = (error) => {
     statusEl.textContent = "Status: Error (Check IP/Firewall)";
     // Re-enable connection controls
   };
   ```

3. **Not Connected**
   ```javascript
   if (ws && ws.readyState === WebSocket.OPEN) {
     // Send command
   } else {
     outputEl.textContent = "Not connected. Please connect first.";
   }
   ```

---

## Flow Diagrams

### Connection Flow

```
User opens remote.html
        ↓
User enters IP and clicks "Connect"
        ↓
JavaScript creates WebSocket
        ↓
HTTP GET /ws (Upgrade: websocket)
        ↓
Server: upgrader.Upgrade()
        ↓
WebSocket connection established
        ↓
Server spawns 2 goroutines:
  - Writer (sends messages)
  - Music poller (every 3s)
        ↓
Client receives player updates every 3s
```

### Command Execution Flow

```
User clicks button
        ↓
sendCommand('command_name') called
        ↓
Check if WebSocket is open
        ↓
Send JSON: {"command": "command_name"}
        ↓
Server receives message
        ↓
Parse JSON → ClientMessage
        ↓
Lookup in ALLOWED_COMMANDS map
        ↓
   Found?
   ├─ Yes → exec.Command().Start()
   │         ↓
   │      Success?
   │      ├─ Yes → Send success response
   │      └─ No  → Send error response
   │
   └─ No  → Send "not allowed" error
        ↓
Client receives response
        ↓
Update output display
```

### Music Update Flow

```
(Every 3 seconds)
        ↓
Ticker fires
        ↓
Call getPlayerInfo()
        ↓
Execute 3 playerctl commands:
  1. metadata --format "..."
  2. status
  3. metadata mpris:artUrl
        ↓
Build info string and artwork URL
        ↓
Send to messages channel
        ↓
Writer goroutine receives
        ↓
conn.WriteJSON(response)
        ↓
Client receives JSON
        ↓
Update player display:
  - Update track text
  - Update album artwork (if available)
```

---

## Performance Considerations

### Resource Usage per Connection

- **Goroutines:** 2 per connection
- **Memory:** ~100KB per connection (rough estimate)
- **CPU:** Minimal, mostly idle (event-driven)

### Scalability

**Current design:**

- Can handle dozens of simultaneous connections
- Each connection independent
- No shared state between connections

**Bottlenecks:**

- playerctl execution (system command overhead)
- JSON marshaling/unmarshaling
- Network I/O

**Optimization opportunities:**

- Cache playerctl results (update once, send to all clients)
- Connection pool limits
- Compression for JSON messages

---

## Extensibility

### Adding New Commands

1. **Add to allowlist:**

   ```go
   var ALLOWED_COMMANDS = map[string][]string{
       "my_command": {"executable", "arg1", "arg2"},
   }
   ```

2. **Add button to HTML:**
   ```html
   <button class="remote-btn" onclick="sendCommand('my_command')">
     🎯 My Command
   </button>
   ```

### Adding Command Output Capture

Currently uses `cmd.Start()` (fire-and-forget). To capture output:

```go
// Instead of:
err := cmd.Start()

// Use:
output, err := cmd.CombinedOutput()  // Captures stdout + stderr
response = ServerResponse{
    Status:  "success",
    Output:  string(output),
}
```

### Adding Authentication

```go
// Environment variable
authToken := os.Getenv("BLITZ_AUTH_TOKEN")

// Validate on each command
if msg.Token != authToken {
    response = ServerResponse{
        Status:  "error",
        Message: "Authentication failed",
    }
    continue
}
```

---

## Testing Strategies

### Manual Testing

1. **Test connection:**

   ```bash
   # Start server
   ./blitz

   # In browser console:
   ws = new WebSocket('ws://localhost:8765/ws')
   ws.onmessage = (e) => console.log(e.data)
   ```

2. **Test command execution:**

   ```javascript
   ws.send(JSON.stringify({ command: "status" }));
   ```

3. **Test playerctl integration:**
   ```bash
   # Play music in Spotify/VLC
   playerctl metadata --format "{{artist}} - {{title}}"
   # Verify in web UI
   ```

### Automated Testing

**Unit test for getPlayerInfo:**

```go
func TestGetPlayerInfo(t *testing.T) {
    info, artwork, err := getPlayerInfo()

    if err != nil && !strings.Contains(info, "not available") {
        t.Errorf("Unexpected error: %v", err)
    }

    // Should return non-empty strings
    if info == "" {
        t.Error("Info should not be empty")
    }
}
```

---

## Troubleshooting Guide

### Server won't start

**Error:** `bind: address already in use`

**Solution:**

```bash
# Find process using port 8765
lsof -i :8765

# Kill the process
kill <PID>
```

### Client can't connect

**Check 1:** Server running?

```bash
ps aux | grep blitz
```

**Check 2:** Firewall blocking?

```bash
sudo ufw status
sudo ufw allow 8765/tcp
```

**Check 3:** Correct IP?

```bash
ip addr  # or ifconfig
```

### Music not showing

**Check 1:** playerctl installed?

```bash
which playerctl
```

**Check 2:** Media player running?

```bash
playerctl status
```

**Check 3:** MPRIS support?

```bash
playerctl -l  # Lists available players
```

### Commands not executing

**Check logs:** Server console will show errors

**Common issues:**

- Command not in allowlist
- Executable not found in PATH
- Permission denied (for sudo commands)

---

## Conclusion

Blitz is a simple yet powerful remote control system built on:

- **WebSocket** for real-time bidirectional communication
- **Go** for efficient, concurrent server implementation
- **Vanilla JavaScript** for lightweight client
- **playerctl** for Linux media player integration

**Strengths:**

- Simple architecture
- Easy to extend
- Low resource usage
- Works on any device with a browser

**Limitations:**

- No authentication
- No encryption
- Limited error handling
- Basic UI

**Perfect for:**

- Controlling your PC from phone/tablet on local network
- Learning WebSocket programming
- Quick home automation tasks

---

_This documentation covers the complete technical implementation of Blitz v0.0.1_
