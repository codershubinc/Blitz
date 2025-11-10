let ws = null;
const statusEl = document.getElementById("status");
const outputEl = document.getElementById("output");
const ipInput = document.getElementById("ipAddress");
const connectBtn = document.getElementById("connectButton");
const playPauseBtn = document.getElementById("play-pause");
const fileInput = document.getElementById("fileInput");
const uploadArea = document.getElementById("uploadArea");
const uploadStatus = document.getElementById("uploadStatus");

// Previous data to detect changes
let previousBluetoothData = null;
let previousWiFiData = null;
let currentServerIP = '';

function updatePlayPauseButton(isPlaying) {
    if (isPlaying) {
        playPauseBtn.textContent = '⏸️';
    } else {
        playPauseBtn.textContent = '▶️';
    }
}

function formatTime(microseconds) {
    if (!microseconds || microseconds === "0") return "0:00";
    const seconds = Math.floor(microseconds / 1000000);
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function updateProgressBar(position, length) {
    const currentTimeEl = document.getElementById('currentTime');
    const totalTimeEl = document.getElementById('totalTime');
    const progressFill = document.getElementById('progressFill');

    if (position && length) {
        const posNum = parseInt(position);
        const lenNum = parseInt(length);
        
        if (lenNum > 0) {
            const percentage = (posNum / lenNum) * 100;
            progressFill.style.width = percentage + '%';
            currentTimeEl.textContent = formatTime(posNum);
            totalTimeEl.textContent = formatTime(lenNum);
        }
    } else {
        progressFill.style.width = '0%';
        currentTimeEl.textContent = '0:00';
        totalTimeEl.textContent = '0:00';
    }
}

function updateBluetoothDevices(devices) {
    const container = document.getElementById('bluetoothDevices');
    
    const currentHash = JSON.stringify(devices.map(d => ({
        name: d.name,
        mac: d.mac,
        battery: d.battery,
        connected: d.connected
    })));

    if (previousBluetoothData === currentHash) {
        return;
    }
    previousBluetoothData = currentHash;

    if (!devices || devices.length === 0) {
        container.innerHTML = '<p class="no-data">No devices connected</p>';
        return;
    }

    container.innerHTML = devices.map(device => `
        <div class="device-card">
            <div class="device-header">
                <span class="device-name">${device.name}</span>
                ${device.battery >= 0 ? `<span class="battery-indicator">${getBatteryIcon(device.battery)} ${device.battery}%</span>` : ''}
            </div>
            <div class="device-details">
                <span class="device-mac">${device.mac}</span>
                <span class="device-icon">${getDeviceIcon(device.icon)}</span>
            </div>
        </div>
    `).join('');
}

function updateWiFiInfo(wifi) {
    const container = document.getElementById('wifiInfo');
    
    const currentHash = JSON.stringify({
        ssid: wifi?.ssid,
        signal: wifi?.signalStrength,
        connected: wifi?.connected,
        download: wifi?.downloadSpeed?.toFixed(2),
        upload: wifi?.uploadSpeed?.toFixed(2),
        ip: wifi?.ipAddress,
        security: wifi?.security
    });

    if (previousWiFiData === currentHash) {
        return;
    }
    previousWiFiData = currentHash;

    if (!wifi || !wifi.connected) {
        container.innerHTML = '<p class="no-data">Not connected</p>';
        return;
    }

    container.innerHTML = `
        <div class="wifi-card">
            <div class="wifi-header">
                <span class="wifi-ssid">${wifi.ssid}</span>
                <span class="wifi-signal">${getSignalIcon(wifi.signalStrength)} ${wifi.signalStrength}%</span>
            </div>
            <div class="wifi-details">
                <div class="speed-info">
                    <span>⬇️ ${wifi.downloadSpeed.toFixed(2)} Mbps</span>
                    <span>⬆️ ${wifi.uploadSpeed.toFixed(2)} Mbps</span>
                </div>
                <div class="wifi-meta">
                    <span class="wifi-freq">${wifi.frequency}</span>
                    ${wifi.security ? `<span class="wifi-security">🔒 ${wifi.security}</span>` : ''}
                    ${wifi.ipAddress ? `<span class="wifi-ip">📍 ${wifi.ipAddress}</span>` : ''}
                    ${wifi.linkSpeed > 0 ? `<span class="wifi-link">🔗 ${wifi.linkSpeed} Mbps</span>` : ''}
                </div>
            </div>
        </div>
    `;
}

function getBatteryIcon(level) {
    if (level >= 75) return '🔋';
    if (level >= 50) return '🔋';
    if (level >= 25) return '🪫';
    return '🪫';
}

function getSignalIcon(strength) {
    if (strength >= 75) return '📶';
    if (strength >= 50) return '📶';
    if (strength >= 25) return '📶';
    return '📶';
}

function getDeviceIcon(iconType) {
    const icons = {
        'audio-card': '🎧',
        'audio-headset': '🎧',
        'audio-headphones': '🎧',
        'input-keyboard': '⌨️',
        'input-mouse': '🖱️',
        'phone': '📱',
        'computer': '💻',
        'video-display': '🖥️'
    };
    return icons[iconType] || '🔵';
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

    currentServerIP = `${ip}:${port}`;
    statusEl.textContent = "Connecting...";
    statusEl.className = "";
    connectBtn.disabled = true;
    ipInput.disabled = true;

    ws = new WebSocket(`ws://${ip}:${port}/ws`);

    ws.onopen = () => {
        console.log("Connected to WebSocket server");
        statusEl.textContent = "Status: Connected";
        statusEl.className = "connected";
        showSkeleton();
    };

    ws.onmessage = (event) => {
        console.log("Message from server:", event.data);
        let data = JSON.parse(event.data);

        if (data.status === 'player') {
            hideSkeleton();
            const playerInfoEl = document.getElementById('playerInfo');
            const albumArtEl = document.getElementById('albumArt');
            const playerControlsEl = document.querySelector('.player-controls');

            const hasActivePlayer = data.output && data.output.Title;

            if (hasActivePlayer) {
                playerControlsEl.classList.remove('hidden');
            } else {
                playerControlsEl.classList.add('hidden');
            }

            if (hasActivePlayer) {
                const info = data.output;
                playerInfoEl.textContent = `${info.Title}\n${info.Artist} - ${info.Album}`;
                updatePlayPauseButton(info.Status === "Playing");
                updateProgressBar(info.Position, info.Length);
            } else {
                playerInfoEl.textContent = 'No music playing';
                updateProgressBar(0, 0);
            }

            if (data.artwork && data.artwork !== '') {
                const artworkChanged = albumArtEl.src !== data.artwork;
                if (artworkChanged) {
                    albumArtEl.classList.add('fade-out');
                    setTimeout(() => {
                        albumArtEl.src = data.artwork;
                        albumArtEl.classList.remove('fade-out');
                        albumArtEl.classList.add('visible');
                    }, 250);
                } else if (!albumArtEl.classList.contains('visible')) {
                    albumArtEl.src = data.artwork;
                    albumArtEl.classList.add('visible');
                }
            } else {
                if (albumArtEl.classList.contains('visible')) {
                    albumArtEl.classList.add('fade-out');
                    setTimeout(() => {
                        albumArtEl.classList.remove('visible', 'fade-out');
                    }, 500);
                }
            }
        } else if (data.status === 'bluetooth') {
            updateBluetoothDevices(data.bluetooth || []);
        } else if (data.status === 'wifi') {
            updateWiFiInfo(data.wifi);
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
        currentServerIP = '';
    };

    ws.onerror = (error) => {
        console.error("WebSocket Error:", error);
        statusEl.textContent = "Status: Error (Check IP/Firewall)";
        statusEl.className = "disconnected";
        connectBtn.disabled = false;
        ipInput.disabled = false;
        ws = null;
        currentServerIP = '';
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

connectBtn.addEventListener('click', connect);
ipInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        connect();
    }
});

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

fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
        uploadFiles(e.target.files);
    }
});

uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.classList.add('drag-over');
});

uploadArea.addEventListener('dragleave', () => {
    uploadArea.classList.remove('drag-over');
});

uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.classList.remove('drag-over');
    if (e.dataTransfer.files.length > 0) {
        uploadFiles(e.dataTransfer.files);
    }
});

async function uploadFiles(files) {
    if (!currentServerIP) {
        uploadStatus.innerHTML = '<p class="error">❌ Not connected to server</p>';
        return;
    }

    uploadStatus.innerHTML = '<p>⏳ Uploading...</p>';

    for (let file of files) {
        await uploadFile(file);
    }
}

async function uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`http://${currentServerIP}/upload`, {
            method: 'POST',
            body: formData
        });

        const result = await response.json();
        
        if (result.status === 'success') {
            uploadStatus.innerHTML += `<p class="success">✅ ${file.name} uploaded successfully</p>`;
        } else {
            uploadStatus.innerHTML += `<p class="error">❌ ${file.name} failed to upload</p>`;
        }
    } catch (error) {
        uploadStatus.innerHTML += `<p class="error">❌ ${file.name} error: ${error.message}</p>`;
    }
}
