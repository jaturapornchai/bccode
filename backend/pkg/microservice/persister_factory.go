package microservice

import "os"

// NewFilePersister เลือก storage backend ตาม config
// ถ้า S3_ENDPOINT ถูก set → ใช้ SeaweedFS S3
// ถ้าไม่ → fallback Azure Blob (backward compatible)
func NewFilePersister() IPersisterFile {
	if os.Getenv("S3_ENDPOINT") != "" {
		return NewPersisterS3()
	}
	return NewPersisterAzureBlob()
}
