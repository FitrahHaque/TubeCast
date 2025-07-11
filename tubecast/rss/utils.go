package rss

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/webp"
)

// Global Variables
var ARCHIVE_IDENTIFIER string //set it at start-up
// var SHAREABLE_FEED_PREFIX = "https://fitrahhaque.github.io/TubeCast"
var SHAREABLE_FEED_PREFIX string
var STATION_BASE string = "tubecast/station"
var StationNames *Set[string]
var FEED_BASE string = "./docs/feed"
var AUDIO_BASE string = "./tubecast/audio"
var COVER_BASE string = "./tubecast/cover"
var THUMBNAIL_BASE string = "./tubecast/thumbnail"
var MaximumStorage uint64 = 2 * 1024 * 1024 * 1024 // 2GB
var Megh Cloud
var Usr User

type FileType int

const (
	THUMBNAIL FileType = iota
	AUDIO
	COVER
	FEED
)

const (
	// Apple Podcasts artwork requirements
	MIN_SIZE = 1400
	MAX_SIZE = 3000
	// DPI      = 72
)

// ImageFormat represents supported output formats
type ImageFormat int

const (
	JPEG ImageFormat = iota
	PNG
)

// var DROPBOX_BASE string =

type Set[T comparable] struct {
	Value map[T]struct{} `json:"value"`
}

// provide key, a comparable type to create a set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		Value: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(item T) {
	s.Value[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.Value, item)
}

func (s *Set[T]) Has(item T) bool {
	_, ok := s.Value[item]
	return ok
}

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	// JSON is expected as e.g. ["a","b","c"]
	var elems []T
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}
	// Allocate the map
	s.Value = make(map[T]struct{}, len(elems))
	// Populate with each element
	for _, e := range elems {
		s.Value[e] = struct{}{}
	}
	return nil
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	// Export as a JSON array of keys
	keys := make([]T, 0, len(s.Value))
	for k := range s.Value {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func run(ctx context.Context, cmd string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	if cmd == "./ia" {
		fmt.Println(cmd, args)
	}
	var out, err bytes.Buffer
	c.Stdout = &out
	c.Stderr = &err
	if e := c.Run(); e != nil {
		return "", fmt.Errorf("could not execute the command. Error: %s\n", &err)
	}
	return out.String(), nil
}

// convertImageForPodcast converts a local image file to JPEG or PNG format
// with proper sizing for Apple Podcasts (1400x1400 to 3000x3000 pixels)
func convertImageForPodcast(inputPath, outputPath string, format ImageFormat, quality int) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		if file, err = os.Open(outputPath); err == nil {
			// fmt.Println("No need for image formatting")
			file.Close()
			return nil
		}
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode the image based on file extension
	var img image.Image
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".webp":
		img, err = webp.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode WebP image: %w", err)
		}
	case ".png":
		img, err = png.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode PNG image: %w", err)
		}
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode JPEG image: %w", err)
		}
	default:
		return fmt.Errorf("unsupported input format: %s", ext)
	}

	// Process the image (resize and make square)
	processedImg := processForPodcast(img)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode based on desired format
	switch format {
	case JPEG:
		options := &jpeg.Options{
			Quality: quality,
		}
		err = jpeg.Encode(outFile, processedImg, options)
		if err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case PNG:
		// fmt.Printf("here I am png\n")
		err = png.Encode(outFile, processedImg)
		if err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output format")
	}

	return nil
}

// processForPodcast resizes and crops image to meet Apple Podcasts requirements
func processForPodcast(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var size int
	if width < height {
		size = width
	} else {
		size = height
	}

	squareImg := cropCenter(img, size, size)

	// Resize to appropriate size for podcasts
	var targetSize uint
	if size < MIN_SIZE {
		targetSize = MIN_SIZE
	} else if size > MAX_SIZE {
		targetSize = MAX_SIZE
	} else {
		targetSize = uint(size)
	}

	resizedImg := resize.Resize(targetSize, targetSize, squareImg, resize.Lanczos3)

	return resizedImg
}

// cropCenter crops an image to the specified dimensions from the center
func cropCenter(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	startX := (imgWidth - width) / 2
	startY := (imgHeight - height) / 2

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(dst, dst.Bounds(), img, image.Pt(startX, startY), draw.Src)

	return dst
}

// getImageDimensions returns the width and height of an image file
func getImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}

// validatePodcastImage checks if an image meets Apple Podcasts requirements
func validatePodcastImage(imagePath string) error {
	width, height, err := getImageDimensions(imagePath)
	if err != nil {
		return err
	}

	if width != height {
		return fmt.Errorf("image must be square, got %dx%d", width, height)
	}

	if width < MIN_SIZE || width > MAX_SIZE {
		return fmt.Errorf("image size must be between %dx%d and %dx%d pixels, got %dx%d",
			MIN_SIZE, MIN_SIZE, MAX_SIZE, MAX_SIZE, width, height)
	}

	return nil
}

func ConvertImageToCorrectFormat(src, dest string) {
	fmt.Printf("src: %v, dest: %v\n", src, dest)
	err := convertImageForPodcast(src, dest, PNG, 0)
	if err != nil {
		fmt.Printf("Error converting to PNG: %v\n", err)
	}
	err = validatePodcastImage(dest)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		// fmt.Println("Image meets Apple Podcasts requirements!")
		os.Remove(src)
		// os.Rename(src, dest)
	}
}

func (user *User) getArchiveIdentifier() string {
	return user.Username + "_tubecast"
}

func (user *User) getFeedUrlPrefix() string {
	return fmt.Sprintf("https://%s.github.io/TubeCast/feed/", user.Username)
}

func (cloud *Cloud) getLocalStationFilepath(title string) string {
	return filepath.Join(STATION_BASE, cloud.getStationFilename(title))
}

func (cloud *Cloud) getLocalFeedFilepath(title string) string {
	return filepath.Join(FEED_BASE, cloud.getFeedFilename(title))
}

func (cloud *Cloud) getLocalCoverFilepath(title string) string {
	return filepath.Join(COVER_BASE, cloud.getCoverFilename(title))
}

func (cloud *Cloud) getLocalThumbnailFilepath(id, title string) string {
	return filepath.Join(THUMBNAIL_BASE, cloud.getThumbnailFilename(id, title))
}

func (cloud *Cloud) getLocalAudioFilepath(id, title string) string {
	return filepath.Join(AUDIO_BASE, cloud.getAudioFilename(id, title))
}

func (cloud *Cloud) getStationFilename(title string) string {
	return title + ".json"
}

func (cloud *Cloud) getFeedFilename(title string) string {
	return fmt.Sprintf("%s.xml", strings.ReplaceAll(title, " ", "_"))
}

func (cloud *Cloud) getCoverFilename(title string) string {
	return fmt.Sprintf("cover_%s.png", title)
}

func (cloud *Cloud) getThumbnailFilename(id string, title string) string {
	return fmt.Sprintf("thumbnail_%s_%s.png", title, id)
}

func (cloud *Cloud) getAudioFilename(id string, title string) string {
	return fmt.Sprintf("audio_%s_%s.mp3", title, id)
}

func (cloud *Cloud) getShareableFeedUrl(title string) string {
	return cloud.FeedUrlPrefix + cloud.getFeedFilename(title)
}

func (cloud *Cloud) getShareableCoverUrl(title string) string {
	return cloud.ArchiveUrlPrefix + cloud.getCoverFilename(title)
}

func (cloud *Cloud) getShareableThumbnailUrl(id, title string) string {
	return cloud.ArchiveUrlPrefix + cloud.getThumbnailFilename(id, title)
}

func (cloud *Cloud) getShareableAudioUrl(id, title string) string {
	return cloud.ArchiveUrlPrefix + cloud.getAudioFilename(id, title)
}
