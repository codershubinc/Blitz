package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ArtworkCache stores album artwork to avoid reading the same file multiple times
// Think of it as a smart memory that remembers images you've already seen
type ArtworkCache struct {
	sync.RWMutex                        // This makes our cache safe for multiple goroutines (concurrent access)
	entries      map[string]*CacheEntry // Map: file path -> cached data
}

// CacheEntry stores one cached image with some extra info to check if it's still valid
type CacheEntry struct {
	dataURI string    // The base64 image data (what the browser needs)
	modTime time.Time // When the file was last changed (to detect updates)
	size    int64     // File size in bytes (another way to detect changes)
}

// NewArtworkCache creates a new empty cache
// Just call this once when your program starts
func NewArtworkCache() *ArtworkCache {
	return &ArtworkCache{
		entries: make(map[string]*CacheEntry), // make() creates an empty map
	}
}

// GetOrFetch is the main function you'll use
// It returns the image as base64, either from cache (fast) or by reading the file (slower)
func (ac *ArtworkCache) GetOrFetch(artworkPath string) (string, error) {
	// Empty path? Nothing to do
	if artworkPath == "" {
		return "", fmt.Errorf("no artwork path provided")
	}

	// Get file info (we need this to check if file has changed)
	fileInfo, err := os.Stat(artworkPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// Try to get from cache first (fast path!)
	if dataURI, found := ac.Get(artworkPath, fileInfo.ModTime(), fileInfo.Size()); found {
		return dataURI, nil // Got it from cache! No need to read file
	}

	// Not in cache, so we need to read the file and convert it
	dataURI, err := ac.readAndEncode(artworkPath)
	if err != nil {
		return "", err
	}

	// Save to cache for next time
	ac.Set(artworkPath, dataURI, fileInfo.ModTime(), fileInfo.Size())

	return dataURI, nil
}

// Get checks if we have this image cached and if it's still valid
// Returns (data, true) if found, ("", false) if not
func (ac *ArtworkCache) Get(path string, modTime time.Time, size int64) (string, bool) {
	ac.RLock()         // Lock for reading (allows multiple readers)
	defer ac.RUnlock() // Unlock when function ends

	entry, exists := ac.entries[path]
	if !exists {
		return "", false // Not in cache
	}

	// Check if file has changed since we cached it
	if entry.modTime.Equal(modTime) && entry.size == size {
		return entry.dataURI, true // Still valid!
	}

	return "", false // File changed, cache is stale
}

// Set saves an image to the cache
func (ac *ArtworkCache) Set(path string, dataURI string, modTime time.Time, size int64) {
	ac.Lock()         // Lock for writing (exclusive access)
	defer ac.Unlock() // Unlock when function ends

	ac.entries[path] = &CacheEntry{
		dataURI: dataURI,
		modTime: modTime,
		size:    size,
	}
}

// readAndEncode reads an image file and converts it to base64
// This is what browsers need to display images
func (ac *ArtworkCache) readAndEncode(artworkPath string) (string, error) {
	// Step 1: Read the entire file into memory as bytes
	imageBytes, err := os.ReadFile(artworkPath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Step 2: Figure out what type of image it is (jpg, png, etc.)
	mimeType := getMimeType(artworkPath)

	// Step 3: Convert bytes to base64 text
	// Base64 is a way to represent binary data (like images) as text
	base64String := base64.StdEncoding.EncodeToString(imageBytes)

	// Step 4: Create a data URI
	// Format: data:image/jpeg;base64,/9j/4AAQSkZJRg...
	// This is what you put in an <img src="..."> tag
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64String)

	return dataURI, nil
}

// getMimeType looks at the file extension and returns the correct MIME type
// MIME type tells the browser what kind of image it is
func getMimeType(path string) string {
	// Get file extension (the part after the last dot)
	ext := filepath.Ext(path) // Example: ".jpg" or ".png"

	// Match extension to MIME type
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg" // If unknown, assume JPEG
	}
}

// Clear removes all cached items (useful for testing)
func (ac *ArtworkCache) Clear() {
	ac.Lock()
	defer ac.Unlock()
	ac.entries = make(map[string]*CacheEntry) // Create fresh empty map
}

// Size returns how many images are cached (useful for debugging)
func (ac *ArtworkCache) Size() int {
	ac.RLock()
	defer ac.RUnlock()
	return len(ac.entries)
}
