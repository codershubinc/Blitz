package utils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func HandleArtworkRequest(artworkPath string) (string, error) {
	// Handle HTTP/HTTPS URLs (download and cache them)
	if strings.HasPrefix(artworkPath, "http://") || strings.HasPrefix(artworkPath, "https://") {
		cachedPath, err := downloadAndCacheArtwork(artworkPath)
		if err != nil {
			return "", err
		}
		artworkPath = cachedPath
	}

	if artworkPath == "" {
		artworkPath = "/home/swap/Downloads/vector-music-note-icon.jpg"
	}

	artworkPath = strings.TrimPrefix(artworkPath, "file://")
	imageBuffer, err := os.ReadFile(artworkPath)
	if err != nil {
		fmt.Println("Something went wrong while reading the file", err)
		return "", err
	}
	// Get file extension and determine image type
	ext := strings.ToLower(filepath.Ext(artworkPath))
	imageExtension := "jpeg" // default

	switch ext {
	case ".png":
		imageExtension = "png"
	case ".jpg", ".jpeg":
		imageExtension = "jpeg"
	case ".gif":
		imageExtension = "gif"
	case ".webp":
		imageExtension = "webp"
	case ".bmp":
		imageExtension = "bmp"
	case ".svg":
		imageExtension = "svg+xml"
	}

	// Return the base64-encoded image data
	return "data:image/" + imageExtension + ";base64," + base64.StdEncoding.EncodeToString(imageBuffer), nil
}

// downloadAndCacheArtwork downloads artwork from URL and caches it locally
func downloadAndCacheArtwork(url string) (string, error) {
	// Extract unique ID from URL
	imageID := extractImageID(url)
	if imageID == "" {
		// Fallback: use MD5 hash of URL
		hash := md5.Sum([]byte(url))
		imageID = fmt.Sprintf("%x", hash)
	}

	// Determine cache directory based on source
	var cacheDir string
	if strings.Contains(url, "scdn.co") || strings.Contains(url, "spotify") {
		cacheDir = "temp/spotify"
	} else {
		cacheDir = "temp/artwork"
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Determine file extension from URL or content-type
	ext := ".jpg"
	if strings.Contains(url, ".png") {
		ext = ".png"
	} else if strings.Contains(url, ".webp") {
		ext = ".webp"
	}

	// Build cache file path
	cachedPath := filepath.Join(cacheDir, imageID+ext)

	// Check if already cached
	if _, err := os.Stat(cachedPath); err == nil {
		return cachedPath, nil
	}

	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download artwork: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download artwork: HTTP %d", resp.StatusCode)
	}

	// Detect content type and adjust extension if needed
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "png") {
		ext = ".png"
		cachedPath = filepath.Join(cacheDir, imageID+ext)
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
		cachedPath = filepath.Join(cacheDir, imageID+ext)
	} else if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		ext = ".jpg"
		cachedPath = filepath.Join(cacheDir, imageID+ext)
	}

	// Create the file
	outFile, err := os.Create(cachedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %v", err)
	}
	defer outFile.Close()

	// Write the image data
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		os.Remove(cachedPath) // Clean up on error
		return "", fmt.Errorf("failed to write image data: %v", err)
	}

	return cachedPath, nil
}

// extractImageID extracts the unique image ID from various CDN URLs
func extractImageID(url string) string {
	// Spotify CDN: https://i.scdn.co/image/ab67616d0000b273270f9f83f24b50fecc041a8d
	// Extract: 270f9f83f24b50fecc041a8d
	spotifyRegex := regexp.MustCompile(`/image/[^/]+([a-f0-9]{24,})`)
	if matches := spotifyRegex.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	// Alternative Spotify pattern: just the hash part
	spotifySimple := regexp.MustCompile(`([a-f0-9]{24,})`)
	if matches := spotifySimple.FindStringSubmatch(url); len(matches) > 0 {
		return matches[0]
	}

	return ""
}
