package microservice

import (
	"context"
)

// consumeBarcodeProjection is no-op in pure PostgreSQL mode
func (ms *Microservice) consumeBarcodeProjection(ctx context.Context, servers, topic, group string, h ServiceHandleFunc) {
	// Synchronous SQL mode: barcode projection runs directly in PostgreSQL
}
