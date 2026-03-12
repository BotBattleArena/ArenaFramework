package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteFrame writes a length-prefixed frame to the writer.
// Format: [4 bytes big-endian uint32 length][N bytes payload]
func WriteFrame(w io.Writer, data []byte) error {
	length := uint32(len(data))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("write frame length: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write frame payload: %w", err)
	}
	return nil
}

// ReadFrame reads a length-prefixed frame from the reader.
// Returns the payload bytes.
func ReadFrame(r io.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("read frame length: %w", err)
	}

	// Sanity check: reject frames larger than 10MB
	const maxFrameSize = 10 * 1024 * 1024
	if length > maxFrameSize {
		return nil, fmt.Errorf("frame too large: %d bytes (max %d)", length, maxFrameSize)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read frame payload: %w", err)
	}
	return buf, nil
}
