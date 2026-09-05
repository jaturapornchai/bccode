// Package projection defines technical delivery metadata and shared PostgreSQL
// lock identities. It does not define money or business lifecycle rules.
package projection

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Metadata struct {
	AggregateUID string `json:"aggregateuid"`
	EventUID     string `json:"eventuid"`
	Version      int64  `json:"version,string"`
}

func AggregateKey(holding, business, guid string) string {
	encoded, _ := json.Marshal([]string{holding, business, guid})
	sum := sha256.Sum256(encoded)
	return "product:" + hex.EncodeToString(sum[:])
}

func WithMetadata(raw json.RawMessage, meta Metadata) (json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if object == nil || meta.AggregateUID == "" || meta.EventUID == "" || meta.Version < 1 {
		return nil, errors.New("invalid projection metadata")
	}
	encoded, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	object["_projection"] = encoded
	return json.Marshal(object)
}

// The shared company lock permits independent rows to run concurrently, while
// a full company rebuild takes the exclusive form before reading MongoDB.
const CompanySharedSQL = "SELECT pg_advisory_xact_lock_shared($1)"
const ExclusiveSQL = "SELECT pg_advisory_xact_lock($1)"

func LockKey(kind, holding, business, code string) int64 {
	encoded, _ := json.Marshal([]string{"bc-projection", kind, holding, business, code})
	sum := sha256.Sum256(encoded)
	return int64(binary.BigEndian.Uint64(sum[:8]))
}
