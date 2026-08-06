package bencode

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Decoder parses bencoded values from a byte slice.
type Decoder struct {
	data []byte
	pos  int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

func (d *Decoder) Position() int {
	return d.pos
}

func (d *Decoder) DecodeInteger() (int, error) {
	if err := d.expect('i'); err != nil {
		return 0, err
	}

	start := d.pos
	for d.pos < len(d.data) && d.data[d.pos] != 'e' {
		d.pos++
	}
	if d.pos >= len(d.data) {
		return 0, errors.New("unterminated integer")
	}
	if start == d.pos {
		return 0, errors.New("empty integer")
	}

	raw := string(d.data[start:d.pos])
	if err := validateInteger(raw); err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", raw, err)
	}
	if value > math.MaxInt || value < math.MinInt {
		return 0, fmt.Errorf("integer %q overflows int", raw)
	}

	d.pos++
	return int(value), nil
}

func (d *Decoder) DecodeString() (string, error) {
	length, err := d.readStringLength()
	if err != nil {
		return "", err
	}
	if d.pos+length > len(d.data) {
		return "", errors.New("string length exceeds input")
	}

	value := string(d.data[d.pos : d.pos+length])
	d.pos += length
	return value, nil
}

func (d *Decoder) DecodeList() ([]any, error) {
	if err := d.expect('l'); err != nil {
		return nil, err
	}

	values := make([]any, 0)
	for {
		if d.pos >= len(d.data) {
			return nil, errors.New("unterminated list")
		}
		if d.data[d.pos] == 'e' {
			d.pos++
			return values, nil
		}

		value, err := d.Decode()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
}

func (d *Decoder) DecodeDictionary() (map[string]any, error) {
	if err := d.expect('d'); err != nil {
		return nil, err
	}

	dict := make(map[string]any)
	for {
		if d.pos >= len(d.data) {
			return nil, errors.New("unterminated dictionary")
		}
		if d.data[d.pos] == 'e' {
			d.pos++
			return dict, nil
		}

		key, err := d.DecodeString()
		if err != nil {
			return nil, fmt.Errorf("invalid dictionary key: %w", err)
		}

		value, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("invalid value for key %q: %w", key, err)
		}
		dict[key] = value
	}
}

func (d *Decoder) Decode() (any, error) {
	if d.pos >= len(d.data) {
		return nil, errors.New("unexpected end of input")
	}

	switch d.data[d.pos] {
	case 'i':
		return d.DecodeInteger()
	case 'l':
		return d.DecodeList()
	case 'd':
		return d.DecodeDictionary()
	default:
		return d.DecodeString()
	}
}

func (d *Decoder) expect(b byte) error {
	if d.pos >= len(d.data) {
		return errors.New("unexpected end of input")
	}
	if d.data[d.pos] != b {
		return fmt.Errorf("expected %q", b)
	}
	d.pos++
	return nil
}

func (d *Decoder) readStringLength() (int, error) {
	if d.pos >= len(d.data) {
		return 0, errors.New("unexpected end of input")
	}
	if d.data[d.pos] < '0' || d.data[d.pos] > '9' {
		return 0, fmt.Errorf("invalid bencode type byte %q", d.data[d.pos])
	}

	start := d.pos
	for d.pos < len(d.data) && d.data[d.pos] != ':' {
		if d.data[d.pos] < '0' || d.data[d.pos] > '9' {
			return 0, errors.New("invalid string length")
		}
		d.pos++
	}
	if d.pos >= len(d.data) {
		return 0, errors.New("unterminated string length")
	}
	if start == d.pos {
		return 0, errors.New("empty string length")
	}
	if d.data[start] == '0' && d.pos-start > 1 {
		return 0, errors.New("string length has leading zero")
	}

	length, err := strconv.ParseInt(string(d.data[start:d.pos]), 10, 0)
	if err != nil {
		return 0, fmt.Errorf("invalid string length: %w", err)
	}
	if length < 0 || length > math.MaxInt {
		return 0, errors.New("string length out of range")
	}

	d.pos++
	return int(length), nil
}

func validateInteger(raw string) error {
	if raw == "-0" {
		return errors.New("negative zero integer is invalid")
	}
	if raw[0] == '-' {
		if len(raw) == 1 {
			return errors.New("integer sign without digits")
		}
		if raw[1] == '0' && len(raw) > 2 {
			return errors.New("integer has leading zero")
		}
		raw = raw[1:]
	} else if raw[0] == '0' && len(raw) > 1 {
		return errors.New("integer has leading zero")
	}

	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return errors.New("invalid integer digit")
		}
	}
	return nil
}
