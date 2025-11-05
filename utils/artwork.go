package utils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func HandleArtworkRequest(artworkPath string) (string, error) {

	if strings.HasPrefix(artworkPath, "http://") || strings.HasPrefix(artworkPath, "https://") {
		return artworkPath, nil
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
