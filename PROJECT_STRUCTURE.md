# Blitz Project Structure Guide

## Current Structure (As-Is)

```
Blitz/
├── main.go                           # Entry point
├── go.mod, go.sum                    # Go modules
├── .env, .env.example                # Environment config
├── README.md                         # Project documentation
├── SPOTIFY_README.md                 # Spotify integration docs
├── blitz                             # Compiled binary (gitignore)
│
├── models/                           # Data models
│   └── server_responce.go            # WebSocket response structure
│
├── utils/                            # Utility functions
│   ├── appLauncher.go                # App launching utilities
│   ├── artwork.go                    # Artwork handling (download/cache)
│   ├── bluetoothInfo.go              # Bluetooth device info
│   ├── mediaInfo.go                  # Media player info (playerctl)
│   ├── spotify.go                    # Spotify API client
│   ├── volumeControls.go             # Volume control utilities
│   ├── wifiInfo.go                   # WiFi status and speed
│   ├── spawnProcesses.go             # Process spawning helper
│   │
│   ├── poller/                       # Polling utilities
│   │   ├── handler.go                # Media poller handler
│   │   └── poller.go                 # Generic polling function
│   │
│   └── websocket/                    # WebSocket handling
│       ├── handler.go                # WebSocket connection handler
│       ├── pingPong.go               # Ping/pong implementation
│       ├── responceChannel.go        # Channel management
│       └── websocke.go               # WebSocket utilities
│
├── web/                              # Frontend files
│   └── index.html                    # WebSocket test client
│
├── temp/                             # Temporary files (gitignore)
│   └── spotify/                      # Cached Spotify artwork
│
└── release/                          # Release binaries
    └── blitz_v0.0.1.2
```

---

## Issues with Current Structure

### 1. **Flat `utils/` Directory**

- ❌ Mixed concerns (player, bluetooth, wifi, spotify)
- ❌ No clear separation between reusable and app-specific code
- ❌ Hard to maintain as project grows

### 2. **Import Cycles**

```
utils/websocket → utils/poller → utils/websocket (CYCLE!)
```

### 3. **Binary in Root**

- ❌ `blitz` binary should be in `bin/` or gitignored

### 4. **No Clear Boundaries**

- ❌ Business logic mixed with utilities
- ❌ Handlers mixed with services

---

## Recommended Structure (Option A: Minimal Changes)

Reorganize `utils/` by domain:

```
Blitz/
├── main.go
├── go.mod, go.sum
├── .env, .env.example
│
├── models/
│   └── response.go              # Renamed from server_responce.go
│
├── utils/
│   ├── player/                  # Media player utilities
│   │   ├── info.go              # Media info (from mediaInfo.go)
│   │   └── controls.go          # Volume controls
│   │
│   ├── bluetooth/               # Bluetooth utilities
│   │   └── device.go            # Device info (from bluetoothInfo.go)
│   │
│   ├── wifi/                    # WiFi utilities
│   │   └── info.go              # WiFi status (from wifiInfo.go)
│   │
│   ├── spotify/                 # Spotify integration
│   │   ├── client.go            # API client (from spotify.go)
│   │   └── artwork.go           # Artwork handling (from artwork.go)
│   │
│   ├── system/                  # System utilities
│   │   ├── launcher.go          # App launcher (from appLauncher.go)
│   │   └── process.go           # Process spawning (from spawnProcesses.go)
│   │
│   ├── poller/                  # Polling utilities
│   │   └── poller.go            # Generic poller (remove handler.go)
│   │
│   └── websocket/               # WebSocket handling
│       ├── handler.go           # Connection handler
│       ├── channel.go           # Channel management
│       ├── connection.go        # Connection utilities
│       └── commands/            # Command handlers
│           └── ping.go          # Ping/pong handler
│
├── bin/                         # Compiled binaries (gitignore)
│   └── blitz
│
├── web/
│   └── index.html
│
└── tmp/                         # Temporary files (gitignore)
    └── spotify/
```

**Benefits:**

- ✅ Clear domain separation
- ✅ Easy to find related code
- ✅ Prevents import cycles
- ✅ Minimal code changes

---

## Recommended Structure (Option B: Full Refactor)

Go standard project layout:

```
Blitz/
├── cmd/                              # Application entrypoints
│   └── blitz/
│       └── main.go                   # Main entry point
│
├── internal/                         # Private application code
│   ├── handlers/                     # Request handlers
│   │   ├── websocket/
│   │   │   ├── handler.go            # WebSocket handler
│   │   │   ├── commands.go           # Command routing
│   │   │   └── writer.go             # Message writer
│   │   │
│   │   └── http/
│   │       └── routes.go             # HTTP routes
│   │
│   ├── services/                     # Business logic
│   │   ├── player/
│   │   │   ├── player.go             # Player service
│   │   │   └── poller.go             # Player poller
│   │   │
│   │   ├── bluetooth/
│   │   │   └── bluetooth.go          # Bluetooth service
│   │   │
│   │   ├── wifi/
│   │   │   └── wifi.go               # WiFi service
│   │   │
│   │   └── spotify/
│   │       ├── client.go             # Spotify client
│   │       └── artwork.go            # Artwork service
│   │
│   ├── models/                       # Internal data models
│   │   ├── response.go               # Response structures
│   │   ├── player.go                 # Player models
│   │   └── device.go                 # Device models
│   │
│   └── config/                       # Configuration
│       └── config.go                 # Config loading
│
├── pkg/                              # Public reusable libraries
│   ├── poller/
│   │   └── poller.go                 # Generic poller
│   │
│   ├── channel/
│   │   └── manager.go                # Channel management
│   │
│   └── process/
│       └── spawn.go                  # Process spawning
│
├── api/                              # API definitions
│   └── websocket/
│       └── messages.go               # WebSocket message types
│
├── web/                              # Frontend assets
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── img/
│   │
│   └── templates/
│       └── index.html
│
├── configs/                          # Configuration files
│   ├── .env.example
│   └── config.yaml
│
├── scripts/                          # Build and deployment scripts
│   ├── build.sh
│   ├── run.sh
│   └── install.sh
│
├── docs/                             # Documentation
│   ├── README.md
│   ├── SPOTIFY.md
│   └── ARCHITECTURE.md
│
├── bin/                              # Compiled binaries (gitignore)
│   └── blitz
│
├── tmp/                              # Temporary files (gitignore)
│   └── spotify/
│
├── .gitignore
├── go.mod
└── go.sum
```

**Benefits:**

- ✅ Standard Go project layout
- ✅ Clear separation of concerns
- ✅ Easy to scale
- ✅ No import cycles possible
- ✅ Public vs private code clearly defined

---

## Import Rules

### Option A (Minimal):

```
main → utils/{domain} ✅
utils/{domain} → models ✅
utils/{domain} ↔ utils/{domain} ✅ (carefully)
```

### Option B (Full):

```
cmd → internal/handlers → internal/services → internal/models ✅
internal/services → pkg ✅
internal → pkg ✅
pkg ← anywhere ✅
internal ← NEVER from outside ❌
```

---

## Migration Path

### Phase 1: Clean Up (1 hour)

```bash
# Move binary
mkdir -p bin
mv blitz bin/
echo "bin/" >> .gitignore
echo "tmp/" >> .gitignore

# Rename temp/ to tmp/
mv temp tmp

# Fix model filename
mv models/server_responce.go models/response.go
```

### Phase 2: Reorganize utils/ (2 hours)

```bash
# Create domain directories
mkdir -p utils/player utils/bluetooth utils/wifi utils/spotify utils/system

# Move files
mv utils/mediaInfo.go utils/player/info.go
mv utils/volumeControls.go utils/player/controls.go
mv utils/bluetoothInfo.go utils/bluetooth/device.go
mv utils/wifiInfo.go utils/wifi/info.go
mv utils/spotify.go utils/spotify/client.go
mv utils/artwork.go utils/spotify/artwork.go
mv utils/appLauncher.go utils/system/launcher.go
mv utils/spawnProcesses.go utils/system/process.go

# Update imports in all files
# (use search/replace or IDE refactoring)
```

### Phase 3: Full Refactor (8 hours)

```bash
# Create new structure
mkdir -p cmd/blitz internal/{handlers,services,models} pkg

# Move main.go
mv main.go cmd/blitz/

# Move and reorganize code
# ... (detailed migration steps)
```

---

## File Organization Patterns

### Good Patterns ✅

**1. Domain-Driven Structure**

```
services/
├── player/
│   ├── player.go      # Main service
│   ├── poller.go      # Player poller
│   └── player_test.go # Tests
```

**2. Feature-Based Structure**

```
handlers/
├── websocket/
│   ├── handler.go     # Main handler
│   ├── commands.go    # Command routing
│   ├── ping.go        # Ping command
│   └── writer.go      # Message writer
```

**3. Layered Structure**

```
internal/
├── handlers/    # Layer 1: Input
├── services/    # Layer 2: Business logic
└── models/      # Layer 3: Data
```

### Bad Patterns ❌

**1. Flat Structure**

```
utils/
├── file1.go
├── file2.go
├── file3.go
... (20 more files)
```

**2. Circular Dependencies**

```
package_a → package_b → package_a (CYCLE!)
```

**3. God Packages**

```
utils/     # Contains everything
common/    # Unclear purpose
helpers/   # Too generic
```

---

## Package Naming Conventions

### Good Names ✅

```go
package player      // Clear domain
package websocket   // Clear purpose
package spotify     // Clear service
```

### Bad Names ❌

```go
package util        // Too generic
package common      // Unclear
package misc        // Everything ends up here
```

---

## Summary

| Option       | Time | Complexity | Benefits                             |
| ------------ | ---- | ---------- | ------------------------------------ |
| **Current**  | 0h   | None       | None - has issues                    |
| **Option A** | 2h   | Low        | Clean domains, fixes most issues     |
| **Option B** | 8h   | High       | Professional, scalable, future-proof |

### Recommendation

1. **Now**: Do Phase 1 (Clean up) - 1 hour
2. **This Week**: Do Option A (Reorganize utils) - 2 hours
3. **Future**: Consider Option B when adding major features

---

## Related Resources

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
