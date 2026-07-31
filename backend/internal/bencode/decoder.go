package bencode

import (
	"errors"
)
type Decoder struct {
	data []byte
	pos  int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data: data,
		pos:  0,
	}
}

func (d *Decoder) DecodeInteger() (int, error) {

	// Integer must start with 'i'
	if d.data[d.pos] != 'i' {
		return 0, errors.New("expected integer")
	}

	// Skip 'i'
	d.pos++

	num := 0

	// Read digits until 'e'
	for d.data[d.pos] != 'e' {

		// Ensure current byte is a digit
		if d.data[d.pos] < '0' || d.data[d.pos] > '9' {
			return 0, errors.New("invalid integer")
		}

		// Build the number
		num = num*10 + int(d.data[d.pos]-'0')

		// Move to next character
		d.pos++
	}

	// Skip 'e'
	d.pos++

	return num, nil
}

func (d *Decoder) DecodeString() (string, error) {

	length := 0

	// Read the length
	for d.data[d.pos] != ':' {

		if d.data[d.pos] < '0' || d.data[d.pos] > '9' {
			return "", errors.New("invalid string length")
		}

		length = length*10 + int(d.data[d.pos]-'0')
		d.pos++
	}

	// Skip ':'
	d.pos++

	// Check if enough bytes are available
	if d.pos+length > len(d.data) {
		return "", errors.New("unexpected end of input")
	}

	// Read the string
	str := string(d.data[d.pos : d.pos+length])

	// Move cursor after the string
	d.pos += length

	return str, nil
}

func (d *Decoder) DecodeList() ([]any, error) {

	// List must start with 'l'
	if d.data[d.pos] != 'l' {
		return nil, errors.New("expected list")
	}

	// Skip 'l'
	d.pos++

	var list []any

	// Read until 'e'
	for d.data[d.pos] != 'e' {

		value, err := d.Decode()
		if err != nil {
			return nil, err
		}

		list = append(list, value)
	}

	// Skip 'e'
	d.pos++

	return list, nil
}

func (d *Decoder) DecodeDictionary() (map[string]any, error) {

	// Dictionary must start with 'd'
	if d.data[d.pos] != 'd' {
		return nil, errors.New("expected dictionary")
	}

	// Skip 'd'
	d.pos++

	dict := make(map[string]any)

	// Read until 'e'
	for d.data[d.pos] != 'e' {

		// Keys are always strings
		key, err := d.DecodeString()
		if err != nil {
			return nil, err
		}

		// Value can be anything
		value, err := d.Decode()
		if err != nil {
			return nil, err
		}

		dict[key] = value
	}

	// Skip 'e'
	d.pos++

	return dict, nil
}

func (d *Decoder) Decode() (any, error) {

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
