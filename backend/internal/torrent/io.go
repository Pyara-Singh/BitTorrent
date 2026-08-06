package torrent

import (
	"fmt"
	"io"
)

func writeFull(w io.Writer, data []byte) error {
	written := 0
	for written < len(data) {
		n, err := w.Write(data[written:])
		written += n
		if err != nil {
			return fmt.Errorf("write failed after %d of %d bytes: %w", written, len(data), err)
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}
