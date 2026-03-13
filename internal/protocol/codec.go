package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Encoder wraps an io.Writer to provide NDJSON encoding.
type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes a value as a JSON string followed by a newline.
func (e *Encoder) Encode(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if _, err := e.w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write json line: %w", err)
	}
	return nil
}

// Decoder wraps an io.Reader to provide NDJSON decoding.
type Decoder struct {
	scanner *bufio.Scanner
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{scanner: bufio.NewScanner(r)}
}

// Decode reads a newline-delimited JSON string and decodes it.
func (d *Decoder) Decode(v interface{}) error {
	if !d.scanner.Scan() {
		if err := d.scanner.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	return json.Unmarshal(d.scanner.Bytes(), v)
}
