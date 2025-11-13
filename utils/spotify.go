package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SpotifyAuth struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"-"`
}

type SpotifyTrack struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Artists    []string `json:"artists"`
	Album      string   `json:"album"`
	AlbumArt   string   `json:"album_art"`
	Duration   int      `json:"duration_ms"`
	Progress   int      `json:"progress_ms"`
	IsPlaying  bool     `json:"is_playing"`
	URI        string   `json:"uri"`
	Popularity int      `json:"popularity"`
}

type SpotifyPlaylist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackCount  int    `json:"track_count"`
	ImageURL    string `json:"image_url"`
	URI         string `json:"uri"`
}

type SpotifyClient struct {
	auth         *SpotifyAuth
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// NewSpotifyClient creates a new Spotify API client
func NewSpotifyClient(clientID, clientSecret, redirectURI string) *SpotifyClient {
	return &SpotifyClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the Spotify authorization URL
func (c *SpotifyClient) GetAuthURL(state string) string {
	scopes := []string{
		"user-read-playback-state",
		"user-modify-playback-state",
		"user-read-currently-playing",
		"playlist-read-private",
		"playlist-read-collaborative",
		"user-library-read",
		"user-top-read",
		"user-read-recently-played",
	}

	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", c.redirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)

	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

// ExchangeCode exchanges authorization code for access token
func (c *SpotifyClient) ExchangeCode(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify auth failed: %s - %s", resp.Status, string(body))
	}

	var auth SpotifyAuth
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return err
	}

	auth.ExpiresAt = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	c.auth = &auth
	return nil
}

// RefreshToken refreshes the access token
func (c *SpotifyClient) RefreshToken() error {
	if c.auth == nil || c.auth.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.auth.RefreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s - %s", resp.Status, string(body))
	}

	var auth SpotifyAuth
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return err
	}

	auth.ExpiresAt = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	// Keep the old refresh token if new one isn't provided
	if auth.RefreshToken == "" {
		auth.RefreshToken = c.auth.RefreshToken
	}
	c.auth = &auth
	return nil
}

// ensureValidToken checks and refreshes token if needed
func (c *SpotifyClient) ensureValidToken() error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated")
	}

	if time.Now().After(c.auth.ExpiresAt.Add(-1 * time.Minute)) {
		return c.RefreshToken()
	}

	return nil
}

// apiRequest makes an authenticated request to Spotify API
func (c *SpotifyClient) apiRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, "https://api.spotify.com/v1"+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.auth.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// GetCurrentTrack gets the currently playing track
func (c *SpotifyClient) GetCurrentTrack() (*SpotifyTrack, error) {
	resp, err := c.apiRequest("GET", "/me/player/currently-playing", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("no track currently playing")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get current track: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Item struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			URI      string `json:"uri"`
			Duration int    `json:"duration_ms"`
			Album    struct {
				Name   string `json:"name"`
				Images []struct {
					URL string `json:"url"`
				} `json:"images"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Popularity int `json:"popularity"`
		} `json:"item"`
		Progress  int  `json:"progress_ms"`
		IsPlaying bool `json:"is_playing"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	track := &SpotifyTrack{
		ID:         result.Item.ID,
		Name:       result.Item.Name,
		Album:      result.Item.Album.Name,
		Duration:   result.Item.Duration,
		Progress:   result.Progress,
		IsPlaying:  result.IsPlaying,
		URI:        result.Item.URI,
		Popularity: result.Item.Popularity,
	}

	// Extract artist names
	for _, artist := range result.Item.Artists {
		track.Artists = append(track.Artists, artist.Name)
	}

	// Get album art URL
	if len(result.Item.Album.Images) > 0 {
		track.AlbumArt = result.Item.Album.Images[0].URL
	}

	return track, nil
}

// Play starts or resumes playback
func (c *SpotifyClient) Play(deviceID string) error {
	endpoint := "/me/player/play"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	resp, err := c.apiRequest("PUT", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("play failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// Pause pauses playback
func (c *SpotifyClient) Pause(deviceID string) error {
	endpoint := "/me/player/pause"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	resp, err := c.apiRequest("PUT", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pause failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// Next skips to next track
func (c *SpotifyClient) Next(deviceID string) error {
	endpoint := "/me/player/next"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	resp, err := c.apiRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("next failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// Previous goes to previous track
func (c *SpotifyClient) Previous(deviceID string) error {
	endpoint := "/me/player/previous"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	resp, err := c.apiRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("previous failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// SetVolume sets the playback volume (0-100)
func (c *SpotifyClient) SetVolume(volume int, deviceID string) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	endpoint := fmt.Sprintf("/me/player/volume?volume_percent=%d", volume)
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	resp, err := c.apiRequest("PUT", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set volume failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetPlaylists gets user's playlists
func (c *SpotifyClient) GetPlaylists(limit int) ([]SpotifyPlaylist, error) {
	endpoint := fmt.Sprintf("/me/playlists?limit=%d", limit)

	resp, err := c.apiRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get playlists failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			URI         string `json:"uri"`
			Tracks      struct {
				Total int `json:"total"`
			} `json:"tracks"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	playlists := make([]SpotifyPlaylist, 0, len(result.Items))
	for _, item := range result.Items {
		playlist := SpotifyPlaylist{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			TrackCount:  item.Tracks.Total,
			URI:         item.URI,
		}
		if len(item.Images) > 0 {
			playlist.ImageURL = item.Images[0].URL
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

// SetAuth sets the authentication manually (useful for loading from storage)
func (c *SpotifyClient) SetAuth(auth *SpotifyAuth) {
	c.auth = auth
}

// GetAuth returns current authentication
func (c *SpotifyClient) GetAuth() *SpotifyAuth {
	return c.auth
}

// IsAuthenticated checks if the client has valid authentication
func (c *SpotifyClient) IsAuthenticated() bool {
	return c.auth != nil && c.auth.AccessToken != ""
}
