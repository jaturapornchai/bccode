package handlers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	webp "github.com/SeriousBug/webp-go-pure/std"
	"golang.org/x/image/draw"
)

const (
	storageImageThumbnailVariant      = "thumbnail"
	storageImageThumbnailMaxDimension = 320
	storageImageThumbnailMaxPixels    = 25_000_000
	storageImageThumbnailQuality      = 82
)

// storageThumbnailObjectKey derives a collision-safe key without changing the
// MongoDB image document. Keeping the original extension in the key also keeps
// foo.jpg and foo.png thumbnails distinct.
func storageThumbnailObjectKey(objectKey string) string {
	objectKey = storageNormalizeObjectKey(objectKey)
	if objectKey == "" {
		return ""
	}
	return objectKey + ".thumb.webp"
}

func storageImageObjectKeys(objectKey string) []string {
	objectKey = storageNormalizeObjectKey(objectKey)
	if objectKey == "" {
		return nil
	}
	return []string{objectKey, storageThumbnailObjectKey(objectKey)}
}

func storageImageVariantObjectKeys(objectKey, variant string) (primary string, fallback string, err error) {
	objectKey = storageNormalizeObjectKey(objectKey)
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "":
		return objectKey, "", nil
	case storageImageThumbnailVariant:
		return storageThumbnailObjectKey(objectKey), objectKey, nil
	default:
		return "", "", fmt.Errorf("unsupported image variant")
	}
}

func createWebPThumbnail(source []byte) ([]byte, error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(source))
	if err != nil {
		return nil, fmt.Errorf("decode image config: %w", err)
	}
	if format != "jpeg" && format != "png" {
		return nil, fmt.Errorf("unsupported image format %q", format)
	}
	if err := validateThumbnailSourceDimensions(config.Width, config.Height); err != nil {
		return nil, err
	}

	sourceImage, decodedFormat, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	if decodedFormat != format {
		return nil, fmt.Errorf("image format changed while decoding")
	}

	width, height := thumbnailDimensions(config.Width, config.Height)
	thumbnail := sourceImage
	if width != config.Width || height != config.Height {
		resized := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(resized, resized.Bounds(), sourceImage, sourceImage.Bounds(), draw.Src, nil)
		thumbnail = resized
	}

	var output bytes.Buffer
	if err := webp.Encode(&output, thumbnail, &webp.Options{
		Quality: storageImageThumbnailQuality,
		Effort:  2,
	}); err != nil {
		return nil, fmt.Errorf("encode WebP thumbnail: %w", err)
	}
	return output.Bytes(), nil
}

func validateThumbnailSourceDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid image dimensions")
	}
	pixels := uint64(width) * uint64(height)
	if pixels > storageImageThumbnailMaxPixels {
		return fmt.Errorf("image has too many pixels")
	}
	return nil
}

func thumbnailDimensions(width, height int) (int, int) {
	if width <= storageImageThumbnailMaxDimension && height <= storageImageThumbnailMaxDimension {
		return width, height
	}
	if width >= height {
		resizedHeight := int((int64(height)*storageImageThumbnailMaxDimension + int64(width)/2) / int64(width))
		if resizedHeight < 1 {
			resizedHeight = 1
		}
		return storageImageThumbnailMaxDimension, resizedHeight
	}
	resizedWidth := int((int64(width)*storageImageThumbnailMaxDimension + int64(height)/2) / int64(height))
	if resizedWidth < 1 {
		resizedWidth = 1
	}
	return resizedWidth, storageImageThumbnailMaxDimension
}
