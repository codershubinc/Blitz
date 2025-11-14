# ⚡ Quazaar

A lightweight WebSocket-based remote control server for Linux systems with real-time music player integration.

## 🎯 Features

- **Remote Command Execution**: Control your PC from any device on your network
- **Real-time Music Display**: Shows currently playing track with album artwork using `playerctl`
- **WebSocket Communication**: Fast, bidirectional communication between devices
- **Secure Command Allowlist**: Only pre-approved commands can be executed
- **Modern Web Interface**: Clean, responsive UI that works on desktop and mobile
- **Auto-updating Music Info**: Track information refreshes every 3 seconds

## 🚀 Quick Start

### Prerequisites

- Go 1.16 or higher
- `playerctl` (for music integration)

  ```bash
  # Arch Linux
  sudo pacman -S playerctl

  # Ubuntu/Debian
  sudo apt install playerctl
  ```

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/codershubinc/Quazaar.git
   cd Quazaar
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Build the server**

   ```bash
   go build -o quazaar
   ```

4. **Run the server**
   ```bash
   ./quazaar
   ```

The server will start on `ws://0.0.0.0:8765/ws`

### Usage

1. **Start the server** on your PC

   ```bash
   ./quazaar
   ```

2. **Open `remote.html`** in a browser on any device

3. **Connect to your PC**

   - Enter your PC's IP address (e.g., `192.168.1.10`)
   - Click "Connect"

4. **Control your PC remotely!**
   - Click buttons to execute commands
   - See currently playing music with album artwork
   - View command outputs in the log area

## 🎵 Music Integration

Quazaar automatically displays your currently playing music using `playerctl`. It shows:

- **Track name** and **artist**
- **Playback status** (Playing/Paused)
- **Album artwork** (when available)

The music info updates every 3 seconds automatically.

### Supported Players

Any media player that supports MPRIS (most Linux media players):

- Spotify
- VLC
- Firefox
- Chrome/Chromium
- Rhythmbox
- And many more!

## ⚙️ Configuration

### Customizing Commands

Edit the `ALLOWED_COMMANDS` map in `main.go` to add or modify commands:

```go
var ALLOWED_COMMANDS = map[string][]string{
    "update":       {"sudo", "pacman", "-Syu"},
    "list_home":    {"ls", "-l", "/home/swap/"},
    "status":       {"git", "status"},
    "open_firefox": {"firefox", "--new-window"},
    "open_edge":    {"microsoft-edge-beta"},
    "open_vscode":  {"code-insiders"},
    "open_postman": {"postman"},
    // Add your custom commands here
    "my_command":   {"command", "arg1", "arg2"},
}
```

**⚠️ Security Note**: Only add commands you trust. This allowlist is your primary security measure.

### Adding Buttons to the UI

Edit `remote.html` and add buttons in the `#remote-grid` div:

```html
<button class="remote-btn" onclick="sendCommand('my_command')">
  🎯 My Command
</button>
```

### Changing the Port

In `main.go`, modify the port in the `main()` function:

```go
err := http.ListenAndServe("0.0.0.0:8765", nil)  // Change 8765 to your port
```

## 🔒 Security Considerations

- **Network Security**: The server listens on all interfaces (`0.0.0.0`). Use firewall rules to restrict access.
- **Command Allowlist**: Only pre-approved commands can be executed. Never add untrusted commands.
- **Authentication**: Currently no authentication. For production use, add proper auth mechanisms.
- **HTTPS**: Uses WebSocket (ws://), not secure WebSocket (wss://). Consider adding TLS for sensitive environments.

### Firewall Configuration (UFW)

```bash
# Allow connections only from your local network
sudo ufw allow from 192.168.1.0/24 to any port 8765

# Or allow from a specific device
sudo ufw allow from 192.168.1.100 to any port 8765
```

## 📱 Mobile Access

The web interface is fully responsive and works great on mobile devices:

1. Make sure your phone is on the same network as your PC
2. Open `remote.html` in your mobile browser
3. Enter your PC's IP address
4. Connect and control!

**Tip**: Bookmark the page or add it to your home screen for quick access.

## 🛠️ Development

### Project Structure

```
Quazaar/
├── main.go         # WebSocket server and command handler
├── remote.html     # Web-based remote control interface
├── go.mod          # Go module dependencies
└── README.md       # This file
```

### Building for Production

go build -ldflags="-s -w" -o quazaar
GOOS=linux GOARCH=amd64 go build -o quazaar-linux-amd64
GOOS=linux GOARCH=arm64 go build -o quazaar-linux-arm64

## 🐛 Troubleshooting

### Server won't start

- Check if port 8765 is already in use: `lsof -i :8765`
- Try a different port

### Can't connect from another device

- Verify firewall settings
- Check that both devices are on the same network
- Use `ip addr` or `ifconfig` to confirm your PC's IP address

### Music info not showing

- Ensure `playerctl` is installed
- Check if a media player is running: `playerctl status`
- Verify MPRIS support in your media player

### Commands not executing

- Check server logs for error messages
- Verify the command exists in `ALLOWED_COMMANDS`
- Ensure the command path is correct

## 📝 License

This project is open source. Feel free to use, modify, and distribute as needed.

## 🤝 Contributing

Contributions are welcome! Feel free to:

- Report bugs
- Suggest new features
- Submit pull requests
- Improve documentation

## 💡 Future enhancement ideas

- [ ] Add authentication/authorization
- [ ] Support for HTTPS/WSS
- [ ] Command history and favorites
- [ ] Multiple user support
- [ ] Custom themes
- [ ] Volume control integration
- [ ] System resource monitoring
- [ ] File transfer capabilities

## 📧 Contact

For questions or feedback, please open an issue on GitHub.

---

**Made with ❤️ By Swapnil Ingle [@codershubinc](https://github.com/codershubinc)**
