let ws = null;
const statusEl = document.getElementById("status");
const outputEl = document.getElementById("output");
const ipInput = document.getElementById("ipAddress");
const connectBtn = document.getElementById("connectButton");
const playPauseBtn = document.getElementById("play-pause");

function updatePlayPauseButton(isPlaying) {
    if (isPlaying) {
        playPauseBtn.textContent = '⏸️'; // Pause icon
    } else {
        playPauseBtn.textContent = '▶️'; // Play icon
    }
}

function connect() {
    const ip = ipInput.value.trim();
    const port = 8765;

    if (!ip) {
        outputEl.textContent = "Please enter an IP address.";
        return;
    }

    if (ws) {
        ws.close();
    }

    statusEl.textContent = "Connecting...";
    statusEl.className = "";
    connectBtn.disabled = true;
    ipInput.disabled = true;

    ws = new WebSocket(`ws://${ip}:${port}/ws`);

    ws.onopen = () => {
        console.log("Connected to WebSocket server");
        statusEl.textContent = "Status: Connected";
        statusEl.className = "connected";

        // Show skeleton while waiting for first update
        showSkeleton();
    };

    ws.onmessage = (event) => {
        console.log("Message from server:", event.data);
        let data = JSON.parse(event.data);

        if (data.status === 'player') {
            // Hide skeleton on first data
            hideSkeleton();

            // Update the player info display
            const playerInfoEl = document.getElementById('playerInfo');
            const albumArtEl = document.getElementById('albumArt');
            const playerControlsEl = document.querySelector('.player-controls');

            // Check if music is actually playing (not "No player running" or "playerctl not available")
            const hasActivePlayer = data.output &&
                !data.output.includes('No player running') &&
                !data.output.includes('playerctl not available') &&
                !data.output.includes('No music');

            // Show/hide player controls
            if (hasActivePlayer) {
                playerControlsEl.classList.remove('hidden');
            } else {
                playerControlsEl.classList.add('hidden');
            }

            // Check if content actually changed
            const textChanged = playerInfoEl.textContent !== (data.output || 'No music playing');
            const artworkChanged = albumArtEl.src !== data.artwork;

            // Update text
            if (textChanged) {
                playerInfoEl.textContent = data.output || 'No music playing';
            }
            // Update play/pause button
            updatePlayPauseButton(data["output"].includes("Playing"));

            // Smooth artwork transition
            if (data.artwork && data.artwork !== '') {
                if (artworkChanged) {
                    // Fade out old image
                    albumArtEl.classList.add('fade-out');

                    setTimeout(() => {
                        albumArtEl.src = data.artwork;
                        albumArtEl.classList.remove('fade-out');
                        albumArtEl.classList.add('visible');
                    }, 250);
                } else if (!albumArtEl.classList.contains('visible')) {
                    // First time showing artwork
                    albumArtEl.src = data.artwork;
                    albumArtEl.classList.add('visible');
                }
            } else {
                // Fade out and hide artwork
                if (albumArtEl.classList.contains('visible')) {
                    albumArtEl.classList.add('fade-out');
                    setTimeout(() => {
                        albumArtEl.classList.remove('visible', 'fade-out');
                    }, 500);
                }
            }
        } else if (data.status === 'success') {
            outputEl.textContent = data.output;
        } else {
            outputEl.textContent = `Error: ${data.message}\n\n${data.output || ''}`;
        }
    };

    ws.onclose = () => {
        console.log("Disconnected from WebSocket server");
        statusEl.textContent = "Status: Disconnected";
        statusEl.className = "disconnected";
        connectBtn.disabled = false;
        ipInput.disabled = false;
        ws = null;
    };

    ws.onerror = (error) => {
        console.error("WebSocket Error:", error);
        statusEl.textContent = "Status: Error (Check IP/Firewall)";
        statusEl.className = "disconnected";
        connectBtn.disabled = false;
        ipInput.disabled = false;
        ws = null;
    };
}

function sendCommand(commandName) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        const message = {
            command: commandName
        };
        ws.send(JSON.stringify(message));
        outputEl.textContent = `Sent command: ${commandName}...`;
    } else {
        outputEl.textContent = "Not connected. Please connect first.";
    }
}

// Add event listeners
connectBtn.addEventListener('click', connect);
ipInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        connect();
    }
});

// Skeleton helper functions
function showSkeleton() {
    document.getElementById('skeletonArt').classList.add('active');
    document.getElementById('skeletonText').classList.add('active');
    document.getElementById('skeletonTextSmall').classList.add('active');
    document.getElementById('playerInfo').classList.add('hidden');
    document.getElementById('albumArt').classList.remove('visible');
}

function hideSkeleton() {
    document.getElementById('skeletonArt').classList.remove('active');
    document.getElementById('skeletonText').classList.remove('active');
    document.getElementById('skeletonTextSmall').classList.remove('active');
    document.getElementById('playerInfo').classList.remove('hidden');
}

