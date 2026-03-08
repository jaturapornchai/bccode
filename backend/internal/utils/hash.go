package utils

import "github.com/segmentio/fasthash/fnv1a"

func FastHash(s string) uint64 {
	return fnv1a.HashString64(s)
}
