package handlers

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	maxProductVideoSizeMB      = int64(500)
	maxProductVideoSize        = maxProductVideoSizeMB * 1024 * 1024
	maxProductVideoRequestSize = maxProductVideoSize + 1024*1024
	productVideoUploadTimeout  = 30 * time.Minute
)

// VideoUploadHandler uploads one private MP4 product or barcode video.
func VideoUploadHandler(c echo.Context) error {
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxProductVideoRequestSize)
	if err := c.Request().ParseMultipartForm(4 * 1024 * 1024); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return c.JSON(http.StatusRequestEntityTooLarge, FileUploadResponse{Success: false, Message: fmt.Sprintf("video is too large; maximum size is %d MB", maxProductVideoSizeMB)})
		}
		return c.JSON(http.StatusBadRequest, FileUploadResponse{Success: false, Message: "invalid video upload"})
	}
	if c.Request().MultipartForm != nil {
		defer c.Request().MultipartForm.RemoveAll()
	}

	holdingCode, authStatus := storageAuthorizedHoldingCode(c, c.FormValue("holdingcode"))
	if authStatus != http.StatusOK {
		return c.JSON(authStatus, FileUploadResponse{Success: false, Message: "shop not selected or forbidden"})
	}
	businessCode, businessStatus := storageAuthorizedBusinessCode(c)
	if businessStatus != http.StatusOK {
		return c.JSON(businessStatus, FileUploadResponse{Success: false, Message: "an active company is required"})
	}

	client, err := GetR2Client()
	if err != nil || client == nil {
		return c.JSON(http.StatusServiceUnavailable, FileUploadResponse{Success: false, Message: storageConfigErrorMessage(err)})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, FileUploadResponse{Success: false, Message: "No video uploaded"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{Success: false, Message: "Failed to process uploaded video"})
	}
	defer src.Close()

	header := make([]byte, 512)
	n, readErr := io.ReadFull(src, header)
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		return c.JSON(http.StatusBadRequest, FileUploadResponse{Success: false, Message: "Failed to read uploaded video"})
	}
	if err := validateProductVideo(file.Filename, file.Size, header[:n]); err != nil {
		return c.JSON(http.StatusBadRequest, FileUploadResponse{Success: false, Message: err.Error()})
	}
	if err := validateMP4Structure(src, file.Size); err != nil {
		return c.JSON(http.StatusBadRequest, FileUploadResponse{Success: false, Message: err.Error()})
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{Success: false, Message: "Failed to process uploaded video"})
	}

	fileName := uuid.NewString() + ".mp4"
	objectKey := fmt.Sprintf("%s/companies/%s/products/videos/%s/%s", holdingCode, storageBusinessPathSegment(businessCode), time.Now().UTC().Format("20060102"), fileName)
	ctx, cancel := context.WithTimeout(c.Request().Context(), productVideoUploadTimeout)
	defer cancel()
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r2BucketName),
		Key:           aws.String(objectKey),
		Body:          src,
		ContentLength: aws.Int64(file.Size),
		ContentType:   aws.String("video/mp4"),
	})
	if err != nil {
		logger.Error("Failed to upload product video: %v", err)
		return c.JSON(http.StatusInternalServerError, FileUploadResponse{Success: false, Message: "Failed to upload video to storage"})
	}

	return c.JSON(http.StatusOK, FileUploadResponse{
		Success:   true,
		FileName:  fileName,
		FileURL:   storageProxyURL(objectKey),
		ObjectKey: objectKey,
		FileSize:  file.Size,
		Message:   "Video uploaded successfully",
	})
}

func validateProductVideo(filename string, size int64, header []byte) error {
	if size <= 0 {
		return fmt.Errorf("video is empty")
	}
	if size > maxProductVideoSize {
		return fmt.Errorf("video is too large; maximum size is %d MB", maxProductVideoSizeMB)
	}
	if strings.ToLower(filepath.Ext(filename)) != ".mp4" {
		return fmt.Errorf("invalid video type; only MP4 is allowed")
	}
	if !hasMP4FileTypeBox(header) {
		return fmt.Errorf("invalid video content; only MP4 is allowed")
	}
	return nil
}

func hasMP4FileTypeBox(header []byte) bool {
	if len(header) < 12 || string(header[4:8]) != "ftyp" {
		return false
	}
	if isMP4Brand(string(header[8:12])) {
		return true
	}
	for offset := 16; offset+4 <= len(header); offset += 4 {
		if isMP4Brand(string(header[offset : offset+4])) {
			return true
		}
	}
	return false
}

func isMP4Brand(brand string) bool {
	switch brand {
	case "isom", "iso2", "avc1", "mp41", "mp42", "M4V ", "dash":
		return true
	default:
		return false
	}
}

func validateMP4Structure(file io.ReadSeeker, size int64) error {
	if size < 24 {
		return fmt.Errorf("invalid video content; MP4 structure is incomplete")
	}
	foundFtyp, foundMoov, foundMdat := false, false, false
	var offset int64
	for boxes := 0; offset < size && boxes < 4096; boxes++ {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("invalid video content; MP4 structure cannot be read")
		}
		var boxHeader [8]byte
		if _, err := io.ReadFull(file, boxHeader[:]); err != nil {
			return fmt.Errorf("invalid video content; MP4 structure is incomplete")
		}
		boxSize := uint64(binary.BigEndian.Uint32(boxHeader[:4]))
		headerSize := uint64(8)
		if boxSize == 1 {
			var extendedSize [8]byte
			if _, err := io.ReadFull(file, extendedSize[:]); err != nil {
				return fmt.Errorf("invalid video content; MP4 structure is incomplete")
			}
			boxSize = binary.BigEndian.Uint64(extendedSize[:])
			headerSize = 16
		} else if boxSize == 0 {
			boxSize = uint64(size - offset)
		}
		if boxSize < headerSize || boxSize > uint64(size-offset) {
			return fmt.Errorf("invalid video content; MP4 box size is invalid")
		}
		switch string(boxHeader[4:8]) {
		case "ftyp":
			foundFtyp = true
		case "moov":
			foundMoov = true
		case "mdat":
			foundMdat = true
		}
		offset += int64(boxSize)
	}
	if offset != size || !foundFtyp || !foundMoov || !foundMdat {
		return fmt.Errorf("invalid video content; MP4 must contain ftyp, moov and mdat")
	}
	return nil
}
