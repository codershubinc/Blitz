# Blitz - Spotify Integration

## Setup Spotify API

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications)
2. Create a new app
3. Add redirect URI: `http://localhost:8765/spotify/callback`
4. Copy your Client ID and Client Secret
5. Create `.env` file from `.env.example`:

   ```bash

   cp .env.example .env

   ```

6. Add your credentials to `.env` file

## Environment Variables

```bash
export SPOTIFY_CLIENT_ID="your_client_id_here"
export SPOTIFY_CLIENT_SECRET="your_client_secret_here"
export SPOTIFY_REDIRECT_URI="http://localhost:8765/spotify/callback"
```

Or source from `.env`:

```bash
export $(cat .env | xargs)
```

## Running with Spotify

```bash
# Set environment variables
export $(cat .env | xargs)

# Build and run
go build -o blitz
./blitz
```

## Authentication Flow

1. Navigate to: `http://localhost:8765/spotify/auth`
2. Login with your Spotify account
3. Grant permissions
4. You'll be redirected back with authentication

## WebSocket Commands

### Authentication Status

```json
{
  "command": "spotify_auth_status"
}
```

### Get Current Track

```json
{
  "command": "spotify_current_track"
}
```

### Playback Controls

```json
{
  "command": "spotify_play",
  "device_id": "optional_device_id"
}

{
  "command": "spotify_pause"
}

{
  "command": "spotify_next"
}

{
  "command": "spotify_previous"
}
```

### Volume Control (0-100)

```json
{
  "command": "spotify_volume",
  "volume": 50,
  "device_id": "optional_device_id"
}
```

### Get Playlists

```json
{
  "command": "spotify_playlists"
}
```

## Response Format

### Current Track

```json
{
  "status": "spotify_track",
  "spotify_track": {
    "id": "track_id",
    "name": "Track Name",
    "artists": ["Artist 1", "Artist 2"],
    "album": "Album Name",
    "album_art": "https://...",
    "duration_ms": 240000,
    "progress_ms": 60000,
    "is_playing": true,
    "uri": "spotify:track:...",
    "popularity": 85
  }
}
```

### Playlists

```json
{
  "status": "spotify_playlists",
  "output": [
    {
      "id": "playlist_id",
      "name": "Playlist Name",
      "description": "Description",
      "track_count": 50,
      "image_url": "https://...",
      "uri": "spotify:playlist:..."
    }
  ]
}
```

## Features

- ✅ OAuth 2.0 authentication flow
- ✅ Automatic token refresh
- ✅ Get currently playing track with full metadata
- ✅ Playback controls (play/pause/next/previous)
- ✅ Volume control
- ✅ Get user playlists
- ✅ Device-specific controls
- ✅ Error handling and status reporting

## Scopes Requested

- `user-read-playback-state` - Read current playback state
- `user-modify-playback-state` - Control playback
- `user-read-currently-playing` - Read currently playing track
- `playlist-read-private` - Read private playlists
- `playlist-read-collaborative` - Read collaborative playlists
- `user-library-read` - Read saved tracks
- `user-top-read` - Read top tracks/artists
- `user-read-recently-played` - Read recently played tracks
