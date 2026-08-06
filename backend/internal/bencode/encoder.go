package bencode

import (
	"bytes"
	"fmt"
	"sort"
)

// Encode writes a canonical bencoded representation of value.
func Encode(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeValue(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, value any) error {
	switch v := value.(type) {
	case int:
		fmt.Fprintf(buf, "i%de", v)
	case int64:
		fmt.Fprintf(buf, "i%de", v)
	case string:
		fmt.Fprintf(buf, "%d:%s", len(v), v)
	case []byte:
		fmt.Fprintf(buf, "%d:", len(v))
		buf.Write(v)
	case []any:
		buf.WriteByte('l')
		for _, item := range v {
			if err := encodeValue(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case map[string]any:
		buf.WriteByte('d')
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(buf, "%d:%s", len(key), key)
			if err := encodeValue(buf, v[key]); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	default:
		return fmt.Errorf("unsupported bencode value type %T", value)
	}
	return nil
}
