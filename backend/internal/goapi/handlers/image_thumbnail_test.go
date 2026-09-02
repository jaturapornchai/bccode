package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	webp "github.com/SeriousBug/webp-go-pure/std"
)

func TestStorageThumbnailObjectKeyPreservesOriginalExtension(t *testing.T) {
	want := "SHOP001/images/photo.jpg.thumb.webp"
	if got := storageThumbnailObjectKey("/SHOP001/images/photo.jpg/"); got != want {
		t.Fatalf("storageThumbnailObjectKey() = %q, want %q", got, want)
	}
}

func TestStorageImageVariantObjectKeys(t *testing.T) {
	primary, fallback, err := storageImageVariantObjectKeys("SHOP001/images/a.png", "thumbnail")
	if err != nil {
		t.Fatalf("storageImageVariantObjectKeys() error = %v", err)
	}
	if primary != "SHOP001/images/a.png.thumb.webp" || fallback != "SHOP001/images/a.png" {
		t.Fatalf("unexpected keys primary=%q fallback=%q", primary, fallback)
	}
	if _, _, err := storageImageVariantObjectKeys("SHOP001/images/a.png", "large"); err == nil {
		t.Fatal("unsupported variant was accepted")
	}
}

func TestStorageImageObjectKeysIncludesOriginalAndThumbnail(t *testing.T) {
	got := storageImageObjectKeys("SHOP001/images/a.png")
	want := []string{"SHOP001/images/a.png", "SHOP001/images/a.png.thumb.webp"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("storageImageObjectKeys() = %v, want %v", got, want)
	}
}

func TestCreateWebPThumbnailResizesAndKeepsAlpha(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 640, 320))
	for y := 0; y < 320; y++ {
		for x := 0; x < 640; x++ {
			alpha := uint8(255)
			if x < 64 && y < 64 {
				alpha = 0
			}
			source.SetNRGBA(x, y, color.NRGBA{R: 25, G: 120, B: 220, A: alpha})
		}
	}

	var pngSource bytes.Buffer
	if err := png.Encode(&pngSource, source); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
	thumbnail, err := createWebPThumbnail(pngSource.Bytes())
	if err != nil {
		t.Fatalf("createWebPThumbnail() error = %v", err)
	}
	if len(thumbnail) < 12 || string(thumbnail[:4]) != "RIFF" || string(thumbnail[8:12]) != "WEBP" {
		t.Fatal("thumbnail is not a WebP container")
	}
	decoded, err := webp.Decode(bytes.NewReader(thumbnail))
	if err != nil {
		t.Fatalf("webp.Decode() error = %v", err)
	}
	if got, want := decoded.Bounds().Size(), image.Pt(320, 160); got != want {
		t.Fatalf("thumbnail size = %v, want %v", got, want)
	}
	_, _, _, alpha := decoded.At(5, 5).RGBA()
	if alpha != 0 {
		t.Fatalf("transparent alpha = %d, want 0", alpha)
	}
}

func TestCreateWebPThumbnailDoesNotUpscale(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	var pngSource bytes.Buffer
	if err := png.Encode(&pngSource, source); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
	thumbnail, err := createWebPThumbnail(pngSource.Bytes())
	if err != nil {
		t.Fatalf("createWebPThumbnail() error = %v", err)
	}
	config, err := webp.DecodeConfig(bytes.NewReader(thumbnail))
	if err != nil {
		t.Fatalf("webp.DecodeConfig() error = %v", err)
	}
	if config.Width != 20 || config.Height != 10 {
		t.Fatalf("thumbnail size = %dx%d, want 20x10", config.Width, config.Height)
	}
}

func TestCreateWebPThumbnailAcceptsJPEG(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 32, 16))
	var jpegSource bytes.Buffer
	if err := jpeg.Encode(&jpegSource, source, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg.Encode() error = %v", err)
	}
	thumbnail, err := createWebPThumbnail(jpegSource.Bytes())
	if err != nil {
		t.Fatalf("createWebPThumbnail() error = %v", err)
	}
	config, err := webp.DecodeConfig(bytes.NewReader(thumbnail))
	if err != nil {
		t.Fatalf("webp.DecodeConfig() error = %v", err)
	}
	if config.Width != 32 || config.Height != 16 {
		t.Fatalf("thumbnail size = %dx%d, want 32x16", config.Width, config.Height)
	}
}

func TestValidateThumbnailSourceDimensionsRejectsDecompressionBombShape(t *testing.T) {
	if err := validateThumbnailSourceDimensions(10_000, 10_000); err == nil {
		t.Fatal("oversized pixel count was accepted")
	}
	if err := validateThumbnailSourceDimensions(6_000, 4_000); err != nil {
		t.Fatalf("ordinary 24MP image was rejected: %v", err)
	}
}
