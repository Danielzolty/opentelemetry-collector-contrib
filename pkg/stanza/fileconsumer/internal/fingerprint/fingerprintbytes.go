// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fingerprint // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/fingerprint"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// Fingerprint is used to identify a file
// A file's fingerprint is the first N bytes of the file
type FingerprintBytes struct {
	FirstBytes []byte
}

// New creates a new fingerprint from an open file
func NewFingerprintBytes(file *os.File, size int) (*Fingerprint, error) {
	buf := make([]byte, size)

	n, err := file.ReadAt(buf, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("reading fingerprint bytes: %w", err)
	}

	var fp Fingerprint = &FingerprintBytes{
		FirstBytes: buf[:n],
	}

	return &fp, nil
}

// Copy creates a new copy of the fingerprint
func (f FingerprintBytes) Copy() *Fingerprint {
	buf := make([]byte, len(f.FirstBytes), cap(f.FirstBytes))
	n := copy(buf, f.FirstBytes)
	var fp Fingerprint = &FingerprintBytes{
		FirstBytes: buf[:n],
	}
	return &fp
}

// Equal returns true if the fingerprints have the same FirstBytes,
// false otherwise. This does not compare other aspects of the fingerprints
// because the primary purpose of a fingerprint is to convey a unique
// identity, and only the FirstBytes field contributes to this goal.
func (f FingerprintBytes) Equal(other Fingerprint) bool {
	l0 := len(other.(FingerprintBytes).FirstBytes)
	l1 := len(f.FirstBytes)
	if l0 != l1 {
		return false
	}
	for i := 0; i < l0; i++ {
		if other.(FingerprintBytes).FirstBytes[i] != f.FirstBytes[i] {
			return false
		}
	}
	return true
}

// StartsWith returns true if the fingerprints are the same
// or if the new fingerprint starts with the old one
// This is important functionality for tracking new files,
// since their initial size is typically less than that of
// a fingerprint. As the file grows, its fingerprint is updated
// until it reaches a maximum size, as configured on the operator
func (f FingerprintBytes) StartsWith(old Fingerprint) bool {
	l0 := len(old.(FingerprintBytes).FirstBytes)
	if l0 == 0 {
		return false
	}
	l1 := len(f.FirstBytes)
	if l0 > l1 {
		return false
	}
	return bytes.Equal(old.(FingerprintBytes).FirstBytes[:l0], f.FirstBytes[:l0])
}

func (f FingerprintBytes) IsMaxSize(maxFingerprintSize int, offset int64) bool {
	return len(f.FirstBytes) == maxFingerprintSize || int(offset) > len(f.FirstBytes)
}

func (f FingerprintBytes) ByteSize() int {
	return len(f.FirstBytes)
}
