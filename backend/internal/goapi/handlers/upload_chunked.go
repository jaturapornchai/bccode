package handlers

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ChunkUploadInfo tracks chunk upload progress
type ChunkUploadInfo struct {
	UploadID       string
	FileName       string
	TotalChunks    int
	ChunkSize      int64
	TotalSize      int64
	UploadedChunks map[int]bool
	Mutex          sync.RWMutex
}

var (
	// Store upload sessions in memory (in production, use Redis or database)
	uploadSessions = make(map[string]*ChunkUploadInfo)
	sessionMutex   sync.RWMutex
)

// InitChunkedUploadHandler initializes a chunked upload session
// POST /upload/init
// JSON body: { "fileName": "example.zip", "totalSize": 1073741824, "chunkSize": 5242880 }
// Returns: { "uploadID": "uuid", "chunkSize": 5242880, "totalChunks": 205 }
func InitChunkedUploadHandler(c echo.Context) error {
	var req struct {
		FileName  string `json:"filename"`
		TotalSize int64  `json:"totalsize"`
		ChunkSize int64  `json:"chunksize"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Error("Invalid request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	// Validate input
	if req.FileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "fileName is required",
		})
	}

	if req.TotalSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "totalSize must be greater than 0",
		})
	}

	// Default chunk size: 5MB
	if req.ChunkSize <= 0 {
		req.ChunkSize = 5 * 1024 * 1024
	}

	// Calculate total chunks
	totalChunks := int((req.TotalSize + req.ChunkSize - 1) / req.ChunkSize)

	// Generate upload ID
	uploadID := uuid.New().String()

	// Create upload session
	uploadInfo := &ChunkUploadInfo{
		UploadID:       uploadID,
		FileName:       req.FileName,
		TotalChunks:    totalChunks,
		ChunkSize:      req.ChunkSize,
		TotalSize:      req.TotalSize,
		UploadedChunks: make(map[int]bool),
	}

	// Store session
	sessionMutex.Lock()
	uploadSessions[uploadID] = uploadInfo
	sessionMutex.Unlock()

	// Create chunks directory in system temp
	tempDir := os.TempDir()
	chunksDir := filepath.Join(tempDir, "uploads", "chunks", uploadID)
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		logger.Error("Failed to create chunks directory: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to create chunks directory",
			"message": err.Error(),
		})
	}

	logger.Info("Chunked upload initialized: uploadID=%s, file=%s, size=%d, chunks=%d",
		uploadID, req.FileName, req.TotalSize, totalChunks)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":     true,
		"uploadID":    uploadID,
		"chunkSize":   req.ChunkSize,
		"totalChunks": totalChunks,
		"message":     "Upload session created successfully",
	})
}

// UploadChunkHandler handles individual chunk uploads
// POST /upload/chunk
// Form-data: uploadID, chunkIndex, chunk (file)
// Returns: { "success": true, "chunkIndex": 0, "progress": 5.5 }
func UploadChunkHandler(c echo.Context) error {
	uploadID := c.FormValue("uploadID")
	chunkIndexStr := c.FormValue("chunkIndex")

	if uploadID == "" || chunkIndexStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "uploadID and chunkIndex are required",
		})
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid chunkIndex",
			"message": err.Error(),
		})
	}

	// Get upload session
	sessionMutex.RLock()
	uploadInfo, exists := uploadSessions[uploadID]
	sessionMutex.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Upload session not found",
		})
	}

	// Validate chunk index
	if chunkIndex < 0 || chunkIndex >= uploadInfo.TotalChunks {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid chunk index",
			"message": fmt.Sprintf("Chunk index must be between 0 and %d", uploadInfo.TotalChunks-1),
		})
	}

	// Get the chunk file
	file, err := c.FormFile("chunk")
	if err != nil {
		logger.Error("Failed to get chunk file: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "No chunk file uploaded",
			"message": err.Error(),
		})
	}

	// Open the chunk file
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open chunk file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to open chunk file",
			"message": err.Error(),
		})
	}
	defer src.Close()

	// Save chunk to disk
	tempDir := os.TempDir()
	chunksDir := filepath.Join(tempDir, "uploads", "chunks", uploadID)
	chunkPath := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d", chunkIndex))

	dst, err := os.Create(chunkPath)
	if err != nil {
		logger.Error("Failed to create chunk file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to create chunk file",
			"message": err.Error(),
		})
	}
	defer dst.Close()

	// Copy chunk data using streaming
	written, err := io.Copy(dst, src)
	if err != nil {
		logger.Error("Failed to save chunk: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to save chunk",
			"message": err.Error(),
		})
	}

	// Mark chunk as uploaded
	uploadInfo.Mutex.Lock()
	uploadInfo.UploadedChunks[chunkIndex] = true
	uploadedCount := len(uploadInfo.UploadedChunks)
	uploadInfo.Mutex.Unlock()

	progress := float64(uploadedCount) / float64(uploadInfo.TotalChunks) * 100

	logger.Debug("Chunk uploaded: uploadID=%s, chunk=%d/%d, size=%d, progress=%.1f%%",
		uploadID, chunkIndex, uploadInfo.TotalChunks-1, written, progress)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":        true,
		"chunkIndex":     chunkIndex,
		"received":       true,
		"bytesWritten":   written,
		"uploadedChunks": uploadedCount,
		"totalChunks":    uploadInfo.TotalChunks,
		"progress":       progress,
	})
}

// MergeChunksHandler merges all uploaded chunks and uploads to Cloudflare R2
// POST /upload/merge
// JSON body: { "uploadID": "uuid", "holdingcode": "optional" }
// Returns: { "success": true, "fileName": "guid.ext", "fileUrl": "presigned", "checksum": "md5hash" }
func MergeChunksHandler(c echo.Context) error {
	var req struct {
		UploadID    string `json:"uploadid"`
		HoldingCode string `json:"holdingcode"`
	}

	if err := c.Bind(&req); err != nil {
		logger.Error("Invalid request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	// Get upload session
	sessionMutex.RLock()
	uploadInfo, exists := uploadSessions[req.UploadID]
	sessionMutex.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Upload session not found",
		})
	}

	// Check if all chunks are uploaded
	uploadInfo.Mutex.RLock()
	uploadedCount := len(uploadInfo.UploadedChunks)
	uploadInfo.Mutex.RUnlock()

	if uploadedCount != uploadInfo.TotalChunks {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Not all chunks uploaded",
			"message": fmt.Sprintf("Uploaded %d/%d chunks", uploadedCount, uploadInfo.TotalChunks),
		})
	}

	logger.Info("Starting merge process: uploadID=%s, file=%s, chunks=%d",
		req.UploadID, uploadInfo.FileName, uploadInfo.TotalChunks)

	// Generate final filename with GUID
	guid := uuid.New().String()
	ext := filepath.Ext(uploadInfo.FileName)
	finalFilename := guid + ext

	// Merge chunks ใน memory buffer
	tempDir := os.TempDir()
	chunksDir := filepath.Join(tempDir, "uploads", "chunks", req.UploadID)

	// Get sorted chunk indices
	chunkIndices := make([]int, 0, uploadInfo.TotalChunks)
	uploadInfo.Mutex.RLock()
	for idx := range uploadInfo.UploadedChunks {
		chunkIndices = append(chunkIndices, idx)
	}
	uploadInfo.Mutex.RUnlock()
	sort.Ints(chunkIndices)

	// Merge chunks into buffer + compute MD5
	var mergedBuf bytes.Buffer
	hash := md5.New()
	multiWriter := io.MultiWriter(&mergedBuf, hash)

	for _, idx := range chunkIndices {
		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d", idx))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			logger.Error("Failed to open chunk file %d: %v", idx, err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to open chunk file",
				"message": err.Error(),
				"chunk":   idx,
			})
		}

		_, err = io.Copy(multiWriter, chunkFile)
		chunkFile.Close()

		if err != nil {
			logger.Error("Failed to merge chunk %d: %v", idx, err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to merge chunk",
				"message": err.Error(),
				"chunk":   idx,
			})
		}
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	totalSize := int64(mergedBuf.Len())

	// Upload merged file ไป Cloudflare R2
	s3Client, err := GetR2Client()
	if err != nil {
		logger.Error("S3 client not available: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Storage service not available",
		})
	}

	holdingCode, authStatus := storageAuthorizedHoldingCode(c, req.HoldingCode)
	if authStatus != http.StatusOK {
		return c.JSON(authStatus, map[string]interface{}{
			"error": "shop not selected or forbidden",
		})
	}
	objectKey := fmt.Sprintf("%s/uploads/%s/%s", holdingCode, time.Now().Format("20060102"), finalFilename)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r2BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(mergedBuf.Bytes()),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		logger.Error("Failed to upload merged file to S3: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to upload merged file to storage",
			"message": err.Error(),
		})
	}

	// สร้าง presigned URL
	fileURL, _ := getPresignedURL(s3Client, objectKey, 60)

	logger.Success("File merged & uploaded to S3: %s (original: %s, size: %d bytes, checksum: %s)",
		finalFilename, uploadInfo.FileName, totalSize, checksum)

	// Cleanup chunks directory
	if err := os.RemoveAll(chunksDir); err != nil {
		logger.Warn("Failed to cleanup chunks directory: %v", err)
	}

	// Remove session
	sessionMutex.Lock()
	delete(uploadSessions, req.UploadID)
	sessionMutex.Unlock()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":      true,
		"message":      "File merged and uploaded successfully",
		"fileName":     finalFilename,
		"fileUrl":      fileURL,
		"objectKey":    objectKey,
		"fileSize":     totalSize,
		"originalName": uploadInfo.FileName,
		"checksum":     checksum,
	})
}

// GetUploadStatusHandler returns the current upload progress
// GET /upload/status/:uploadID
// Returns: { "uploadID": "uuid", "progress": 75.5, "uploadedChunks": 15, "totalChunks": 20 }
func GetUploadStatusHandler(c echo.Context) error {
	uploadID := c.Param("uploadID")

	if uploadID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "uploadID is required",
		})
	}

	// Get upload session
	sessionMutex.RLock()
	uploadInfo, exists := uploadSessions[uploadID]
	sessionMutex.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Upload session not found",
		})
	}

	uploadInfo.Mutex.RLock()
	uploadedCount := len(uploadInfo.UploadedChunks)
	uploadedChunks := make([]int, 0, uploadedCount)
	for idx := range uploadInfo.UploadedChunks {
		uploadedChunks = append(uploadedChunks, idx)
	}
	uploadInfo.Mutex.RUnlock()

	sort.Ints(uploadedChunks)
	progress := float64(uploadedCount) / float64(uploadInfo.TotalChunks) * 100

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":        true,
		"uploadID":       uploadID,
		"fileName":       uploadInfo.FileName,
		"totalSize":      uploadInfo.TotalSize,
		"chunkSize":      uploadInfo.ChunkSize,
		"totalChunks":    uploadInfo.TotalChunks,
		"uploadedChunks": uploadedCount,
		"uploadedList":   uploadedChunks,
		"progress":       progress,
		"isComplete":     uploadedCount == uploadInfo.TotalChunks,
	})
}

// CancelUploadHandler cancels an upload session and cleans up chunks
// DELETE /upload/cancel/:uploadID
// Returns: { "success": true, "message": "Upload cancelled" }
func CancelUploadHandler(c echo.Context) error {
	uploadID := c.Param("uploadID")

	if uploadID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "uploadID is required",
		})
	}

	// Get upload session
	sessionMutex.Lock()
	uploadInfo, exists := uploadSessions[uploadID]
	if exists {
		delete(uploadSessions, uploadID)
	}
	sessionMutex.Unlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Upload session not found",
		})
	}

	// Cleanup chunks directory
	tempDir := os.TempDir()
	chunksDir := filepath.Join(tempDir, "uploads", "chunks", uploadID)
	if err := os.RemoveAll(chunksDir); err != nil {
		logger.Warn("Failed to cleanup chunks directory: %v", err)
	}

	logger.Info("Upload cancelled: uploadID=%s, file=%s", uploadID, uploadInfo.FileName)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Upload cancelled: %s", uploadInfo.FileName),
	})
}
