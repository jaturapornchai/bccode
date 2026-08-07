package handlers

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestValidateProductVideo(t *testing.T) {
	mp4 := append(mp4Box("ftyp", []byte{'i', 's', 'o', 'm', 0, 0, 2, 0, 'i', 's', 'o', 'm', 'a', 'v', 'c', '1'}), mp4Box("moov", nil)...)
	mp4 = append(mp4, mp4Box("mdat", []byte{0})...)
	if err := validateProductVideo("product.mp4", int64(len(mp4)), mp4); err != nil {
		t.Fatalf("valid MP4 header was rejected: %v", err)
	}
	if err := validateMP4Structure(bytes.NewReader(mp4), int64(len(mp4))); err != nil {
		t.Fatalf("valid MP4 structure was rejected: %v", err)
	}
	if err := validateProductVideo("renamed.mp4", 4, []byte("noop")); err == nil {
		t.Fatal("renamed non-MP4 content was accepted")
	}
	if err := validateProductVideo("product.webm", int64(len(mp4)), mp4); err == nil {
		t.Fatal("non-MP4 extension was accepted")
	}
	if err := validateProductVideo("empty.mp4", 0, mp4); err == nil {
		t.Fatal("empty video was accepted")
	}
	if err := validateProductVideo("over-50mb.mp4", 51*1024*1024, mp4); err != nil {
		t.Fatalf("video over 50 MB but within the configured limit was rejected: %v", err)
	}
	if err := validateProductVideo("large.mp4", maxProductVideoSize+1, mp4); err == nil {
		t.Fatal("oversized video was accepted")
	}
	ftypOnly := mp4Box("ftyp", []byte{'i', 's', 'o', 'm'})
	if err := validateMP4Structure(bytes.NewReader(ftypOnly), int64(len(ftypOnly))); err == nil {
		t.Fatal("MP4 without moov and mdat was accepted")
	}
}

func mp4Box(boxType string, payload []byte) []byte {
	box := make([]byte, 8+len(payload))
	binary.BigEndian.PutUint32(box[:4], uint32(len(box)))
	copy(box[4:8], boxType)
	copy(box[8:], payload)
	return box
}
