// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package reader // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/reader"

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/emit"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/fingerprint"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/header"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/internal/scanner"
	"hash/fnv"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/decode"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/flush"
)

type Config struct {
	FingerprintSize         int
	MaxLogSize              int
	Emit                    emit.Callback
	IncludeFileName         bool
	IncludeFilePath         bool
	IncludeFileNameResolved bool
	IncludeFilePathResolved bool
	DeleteAtEOF             bool
	FlushTimeout            time.Duration
}

type Metadata struct {
	Fingerprint     *fingerprint.Fingerprint
	Offset          int64
	FileAttributes  map[string]any
	HeaderFinalized bool
	FlushState      *flush.State
}

// Reader manages a single file
type Reader struct {
	*Config
	*Metadata
	featureGate   bool
	fileName      string
	logger        *zap.SugaredLogger
	file          *os.File
	lineSplitFunc bufio.SplitFunc
	splitFunc     bufio.SplitFunc
	decoder       *decode.Decoder
	headerReader  *header.Reader
	processFunc   emit.Callback
}

func (r *Reader) SetFeatureGate() {
	r.featureGate = true
}

// offsetToEnd sets the starting offset
func (r *Reader) offsetToEnd() error {
	info, err := r.file.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	r.Offset = info.Size()
	return nil
}

func (r *Reader) NewFingerprintFromFile() (*fingerprint.Fingerprint, error) {
	if r.file == nil {
		return nil, errors.New("file is nil")
	}
	if r.featureGate == true {
		return fingerprint.NewFingerprintHash(r.file, r.FingerprintSize)
	}
	return fingerprint.NewFingerprintBytes(r.file, r.FingerprintSize)
}

// ReadToEnd will read until the end of the file
func (r *Reader) ReadToEnd(ctx context.Context) {
	if _, err := r.file.Seek(r.Offset, 0); err != nil {
		r.logger.Errorw("Failed to seek", zap.Error(err))
		return
	}

	s := scanner.New(r, r.MaxLogSize, scanner.DefaultBufferSize, r.Offset, r.splitFunc)

	// Iterate over the tokenized file, emitting entries as we go
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ok := s.Scan()
		if !ok {
			if err := s.Error(); err != nil {
				r.logger.Errorw("Failed during scan", zap.Error(err))
			} else if r.DeleteAtEOF {
				r.delete()
			}
			break
		}

		token, err := r.decoder.Decode(s.Bytes())
		if err != nil {
			r.logger.Errorw("decode: %w", zap.Error(err))
		} else if err := r.processFunc(ctx, token, r.FileAttributes); err != nil {
			if errors.Is(err, header.ErrEndOfHeader) {
				r.finalizeHeader()

				// Now that the header is consumed, use the normal split and process functions.
				// Recreate the scanner with the normal split func.
				// Do not use the updated offset from the old scanner, as the most recent token
				// could be split differently with the new splitter.
				r.splitFunc = r.lineSplitFunc
				r.processFunc = r.Emit
				if _, err = r.file.Seek(r.Offset, 0); err != nil {
					r.logger.Errorw("Failed to seek post-header", zap.Error(err))
					return
				}
				s = scanner.New(r, r.MaxLogSize, scanner.DefaultBufferSize, r.Offset, r.splitFunc)
			} else {
				r.logger.Errorw("process: %w", zap.Error(err))
			}
		}
		r.Offset = s.Pos()
	}
}

func (r *Reader) finalizeHeader() {
	if err := r.headerReader.Stop(); err != nil {
		r.logger.Errorw("Failed to stop header pipeline during finalization", zap.Error(err))
	}
	r.headerReader = nil
	r.HeaderFinalized = true
}

// Delete will close and delete the file
func (r *Reader) delete() {
	r.Close()
	if err := os.Remove(r.fileName); err != nil {
		r.logger.Errorf("could not delete %s", r.fileName)
	}
}

// Close will close the file and return the metadata
func (r *Reader) Close() *Metadata {
	if r.file != nil {
		if err := r.file.Close(); err != nil {
			r.logger.Debugw("Problem closing reader", zap.Error(err))
		}
		r.file = nil
	}

	if r.headerReader != nil {
		if err := r.headerReader.Stop(); err != nil {
			r.logger.Errorw("Failed to stop header pipeline", zap.Error(err))
		}
	}
	m := r.Metadata
	r.Metadata = nil
	return m
}

// Read from the file and update the fingerprint if necessary
func (r *Reader) Read(dst []byte) (int, error) {
	// Skip if fingerprint is already built
	// or if fingerprint is behind Offset
	if r.featureGate == true {
		var fp fingerprint.FingerprintHash = (*r.Fingerprint).(fingerprint.FingerprintHash)
		if fp.IsMaxSize(r.FingerprintSize, r.Offset) {
			return r.file.Read(dst)
		}
		n, err := r.file.Read(dst)
		appendCount := min0(n, r.FingerprintSize-int(r.Offset))
		// return for n == 0 or r.Offset >= r.FingerprintSize
		if appendCount == 0 {
			return n, err
		}

		// for appendCount==0, the following code would add `0` to fingerprint
		if fp.FirstBytes == nil {
			fp.FirstBytes = dst[:appendCount]
		} else {
			fp.FirstBytes = append(fp.FirstBytes[:r.Offset], dst[:appendCount]...)
		}
		if fp.HashInstance == nil {
			h := fnv.New64()
			h.Write(fp.FirstBytes)
			hashed := h.Sum64()
			fp.HashInstance = &h
			fp.HashBytes = hashed
		} else {
			hashInstance := *fp.HashInstance
			hashInstance.Write(fp.FirstBytes[fp.BytesUsed:len(fp.FirstBytes)])
			fp.HashBytes = hashInstance.Sum64()
		}
		fp.BytesUsed = len(fp.FirstBytes)

		return n, err
	} else {
		var fp fingerprint.FingerprintBytes = (*r.Fingerprint).(fingerprint.FingerprintBytes)
		if len(fp.FirstBytes) == r.FingerprintSize || int(r.Offset) > len(fp.FirstBytes) {
			return r.file.Read(dst)
		}
		n, err := r.file.Read(dst)
		appendCount := min0(n, r.FingerprintSize-int(r.Offset))
		// return for n == 0 or r.Offset >= r.fingerprintSize
		if appendCount == 0 {
			return n, err
		}

		// for appendCount==0, the following code would add `0` to fingerprint
		fp.FirstBytes = append(fp.FirstBytes[:r.Offset], dst[:appendCount]...)
		return n, err
	}

}

func min0(a, b int) int {
	if a < 0 || b < 0 {
		return 0
	}
	if a < b {
		return a
	}
	return b
}

func (r *Reader) NameEquals(other *Reader) bool {
	return r.fileName == other.fileName
}

// Validate returns true if the reader still has a valid file handle, false otherwise.
func (r *Reader) Validate() bool {
	if r.file == nil {
		return false
	}
	if r.featureGate == true {
		refreshedFingerprint, err := fingerprint.NewFingerprintHash(r.file, r.FingerprintSize)
		if err != nil {
			return false
		}
		var fp fingerprint.FingerprintHash = (*refreshedFingerprint).(fingerprint.FingerprintHash)
		if fp.StartsWith(*(r.Fingerprint)) {
			return true
		}
		return false
	} else {
		refreshedFingerprint, err := fingerprint.NewFingerprintBytes(r.file, r.FingerprintSize)
		if err != nil {
			return false
		}
		var fp fingerprint.FingerprintBytes = (*refreshedFingerprint).(fingerprint.FingerprintBytes)
		if fp.StartsWith(*(r.Fingerprint)) {
			return true
		}
		return false
	}

}
