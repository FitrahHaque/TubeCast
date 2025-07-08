package rss

import (
	"bytes"
	"context"
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
var STATION_BASE string = "tubecast/stations"
var StationNames *Set[string]
var THUMBNAILS_BASE string = "tubecast/thumbnails"
var AUDIO_BASE string = "tubecast/audio"
var FEED_BASE string = "tubecast/feed"
var COVER_BASE string = "tubecast/cover"
var DROPBOX_AUDIO_BASE string = "/PodcastAudio"
var DROPBOX_THUMBNAILS_BASE string = "/PodcastThumbnails"
var DROPBOX_FEED_BASE string = "/PodcastRSSFeed"
var DROPBOX_COVER_BASE string = "/PodcastCover"
var MaximumCloudStorage uint64 = 2 * 1024 * 1024 * 1024
var TOKEN_MANAGER *TokenManager

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
	DPI      = 72
)

var dpi float64 = 72.0
var inchToMeter float64 = 39.37007874 // inches per meter
var physChunkName string = "pHYs"
var pngSignature string = "\x89PNG\r\n\x1a\n"

// ImageFormat represents supported output formats
type ImageFormat int

const (
	JPEG ImageFormat = iota
	PNG
)

// var DROPBOX_BASE string =

type Set[T comparable] struct {
	set map[T]struct{}
}

// provide key, a comparable type to create a set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		set: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(item T) {
	s.set[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.set, item)
}

func (s *Set[T]) Has(item T) bool {
	_, ok := s.set[item]
	return ok
}

func run(ctx context.Context, cmd string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
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
			fmt.Println("No need for image formatting")
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

	// Make the image square by cropping to the smaller dimension
	var size int
	if width < height {
		size = width
	} else {
		size = height
	}

	// Create a square crop from the center
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

	// Resize using high-quality Lanczos resampling
	resizedImg := resize.Resize(targetSize, targetSize, squareImg, resize.Lanczos3)

	return resizedImg
}

// cropCenter crops an image to the specified dimensions from the center
func cropCenter(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Calculate crop coordinates
	startX := (imgWidth - width) / 2
	startY := (imgHeight - height) / 2

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Crop using draw package
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

	// Check if image is square
	if width != height {
		return fmt.Errorf("image must be square, got %dx%d", width, height)
	}

	// Check size requirements
	if width < MIN_SIZE || width > MAX_SIZE {
		return fmt.Errorf("image size must be between %dx%d and %dx%d pixels, got %dx%d",
			MIN_SIZE, MIN_SIZE, MAX_SIZE, MAX_SIZE, width, height)
	}

	return nil
}

func ConvertImageToCorrectFormat(base_path, name string) {
	src := filepath.Join(base_path, name+".webp")
	dest := filepath.Join(base_path, name+".png")
	err := convertImageForPodcast(src, dest, PNG, 0)
	if err != nil {
		fmt.Printf("Error converting to PNG: %v\n", err)
	} else {
		fmt.Println("Successfully converted to PNG")
	}
	// if err = setPNG72DPI(dest); err != nil {
	// 	fmt.Printf("error: %v\n", err)
	// }
	// Example 3: Validate the output
	err = validatePodcastImage(dest)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Image meets Apple Podcasts requirements!")
		os.Remove(src)
		// os.Rename(src, dest)
	}
}

// // setPNG72DPI reads infile, sets or replaces its pHYs chunk to 72 DPI, and writes to outfile.
// // If outfile is empty, it overwrites infile.
// func setPNG72DPI(infile string) error {
// 	outfile := infile + ".tmp"
// 	defer os.Remove(outfile)
// 	// Open input
// 	in, err := os.Open(infile)
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()

// 	// Create output
// 	out, err := os.Create(outfile)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()

// 	// Write PNG signature
// 	sig := make([]byte, len(pngSignature))
// 	if _, err := io.ReadFull(in, sig); err != nil {
// 		return fmt.Errorf("reading PNG signature: %w", err)
// 	}
// 	if string(sig) != pngSignature {
// 		return fmt.Errorf("invalid PNG signature")
// 	}
// 	if _, err := out.Write(sig); err != nil {
// 		return err
// 	}

// 	// Compute pixels-per-unit for 72 DPI
// 	ppum := uint32(dpi*inchToMeter + 0.5)

// 	// Copy chunks, inserting pHYs after IHDR and skipping any existing pHYs
// 	for {
// 		// Read chunk length (4 bytes)
// 		var length uint32
// 		if err := binary.Read(in, binary.BigEndian, &length); err != nil {
// 			return fmt.Errorf("reading chunk length: %w", err)
// 		}
// 		// Read chunk type (4 bytes)
// 		typ := make([]byte, 4)
// 		if _, err := io.ReadFull(in, typ); err != nil {
// 			return fmt.Errorf("reading chunk type: %w", err)
// 		}

// 		// Read chunk data and CRC
// 		data := make([]byte, length)
// 		if _, err := io.ReadFull(in, data); err != nil {
// 			return fmt.Errorf("reading chunk data: %w", err)
// 		}
// 		crc := make([]byte, 4)
// 		if _, err := io.ReadFull(in, crc); err != nil {
// 			return fmt.Errorf("reading chunk CRC: %w", err)
// 		}

// 		// Write IHDR, then insert new pHYs
// 		if string(typ) == "IHDR" {
// 			// Write IHDR as-is
// 			if err := writeChunk(out, typ, data); err != nil {
// 				return err
// 			}
// 			// Build and write pHYs chunk
// 			physData := make([]byte, 9)
// 			binary.BigEndian.PutUint32(physData[0:4], ppum)
// 			binary.BigEndian.PutUint32(physData[4:8], ppum)
// 			physData[8] = 1 // unit: meter
// 			if err := writeChunk(out, []byte(physChunkName), physData); err != nil {
// 				return err
// 			}
// 		} else if string(typ) == physChunkName {
// 			// Skip existing pHYs chunk
// 			// do nothing
// 		} else {
// 			// Write other chunks unchanged
// 			if err := writeChunk(out, typ, data); err != nil {
// 				return err
// 			}
// 		}

// 		// Stop at IEND
// 		if string(typ) == "IEND" {
// 			break
// 		}
// 	}

// 	if err := os.Rename(outfile, infile); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // writeChunk writes a single PNG chunk (length, type, data, crc) to w.
// func writeChunk(w io.Writer, typ, data []byte) error {
// 	length := uint32(len(data))
// 	// Write length
// 	if err := binary.Write(w, binary.BigEndian, length); err != nil {
// 		return err
// 	}
// 	// Write type and data
// 	if _, err := w.Write(typ); err != nil {
// 		return err
// 	}
// 	if _, err := w.Write(data); err != nil {
// 		return err
// 	}
// 	// Compute and write CRC
// 	crc := crc32.NewIEEE()
// 	crc.Write(typ)
// 	crc.Write(data)
// 	if err := binary.Write(w, binary.BigEndian, crc.Sum32()); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func embedCover(name string) error {
// 	mp3path := filepath.Join(AUDIO_BASE, name+".mp3")
// 	imgpath := filepath.Join(THUMBNAILS_BASE, name+".png")
// 	tag, err := id3v2.Open(mp3path, id3v2.Options{Parse: true})
// 	if err != nil {
// 		return fmt.Errorf("error opening mp3 file: %w", err)
// 	}
// 	defer tag.Close()

// 	imgData, err := os.ReadFile(imgpath)
// 	if err != nil {
// 		return fmt.Errorf("error reading image file: %w", err)
// 	}

// 	// Set the correct MIME type for your image
// 	mimeType := "image/png"
// 	// if len(imgData) > 4 && imgData[1] == 'P' && imgData[2] == 'N' && imgData[3] == 'G' {
// 	// 	mimeType = "image/png"
// 	// }

// 	pic := id3v2.PictureFrame{
// 		Encoding:    id3v2.EncodingUTF8,
// 		MimeType:    mimeType,
// 		PictureType: id3v2.PTFrontCover,
// 		Description: "Episode cover",
// 		Picture:     imgData,
// 	}
// 	tag.AddAttachedPicture(pic)

// 	if err := tag.Save(); err != nil {
// 		return fmt.Errorf("error saving tag: %w", err)
// 	}
// 	return nil
// }
