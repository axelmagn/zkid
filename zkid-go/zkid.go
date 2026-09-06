package zkid

import (
	"errors"
	"fmt"
	// "math"
	// "strconv"
	"strings"
	"time"
	// "unicode"
	// "unicode/utf8"
)

var (
	ErrInvalidFormat    = errors.New("invalid zkid format")
	ErrInvalidTimestamp = errors.New("invalid zkid timestamp")
	ErrInvalidSeparator = errors.New("invalid separator character")
	ErrYearOutOfRange   = errors.New("year out of range (must be 1..9999)")
	ErrDateOutOfRange   = errors.New("date or time component out of range")
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var decodeBase62Table [256]int8

// build decode table from alphabet
func init() {
	for i := range decodeBase62Table {
		decodeBase62Table[i] = -1
	}
	for i := 0; i < len(base62Alphabet); i++ {
		decodeBase62Table[base62Alphabet[i]] = int8(i)
	}
}

// ZkidFormat defines parameters for encoding and decoding ZKID timestamp slugs.
type ZkidFormat struct {
	Century        uint16 // Epoch century
	Separator      byte   // Separator character
	SeparatorDepth uint8  // Separator depth (from hour mark)
}

// DefaultFormat is the default ZKID format (zkid 20-1).
var DefaultFormat = ZkidFormat{
	Century:        20,
	Separator:      '-',
	SeparatorDepth: 1,
}

func encodeBase62(val byte) (byte, error) {
	if int(val) >= len(base62Alphabet) {
		return 0, ErrInvalidTimestamp
	}
	return base62Alphabet[val], nil
}

func decodeBase62(b byte) (byte, error) {
	v := decodeBase62Table[b]
	if v < 0 {
		return 0, ErrInvalidTimestamp
	}
	return byte(v), nil
}

// Encode encodes a time.Time into a ZKID string using the given format.
func Encode(t time.Time, format ZkidFormat, minYearWidth uint8, rightPrecision uint8) (string, error) {
	// TODO: input validation

	year := t.Year()
	// TODO: assert year <= 9999
	yearStr := fmt.Sprintf("%d", year)

	epoch := format.Century * 100
	epochStr := fmt.Sprintf("%d", epoch)

	// pad year and epoch strings to same length
	yearMaxWidth := max(len(yearStr), len(epochStr), int(minYearWidth))
	if len(yearStr) < yearMaxWidth {
		var sb strings.Builder
		for range yearMaxWidth - len(yearStr) {
			sb.WriteByte('0')
		}
		sb.WriteString(yearStr)
		yearStr = sb.String()
	}
	if len(epochStr) < yearMaxWidth {
		var sb strings.Builder
		for range yearMaxWidth - len(epochStr) {
			sb.WriteByte('0')
		}
		sb.WriteString(epochStr)
		epochStr = sb.String()
	}

	// slice year by omitting digits shared with epoch
	if int(minYearWidth) < yearMaxWidth {
		yearSliceIdx := yearMaxWidth - int(minYearWidth)
		for i := range yearSliceIdx {
			if yearStr[i] != epochStr[i] {
				yearSliceIdx = i
				break
			}
		}
		yearStr = yearStr[yearSliceIdx:]
	}

	monthChar, err := encodeBase62(byte(t.Month()))
	if err != nil {
		return "", err
	}

	dayChar, err := encodeBase62(byte(t.Day()))
	if err != nil {
		return "", err
	}

	hourChar, err := encodeBase62(byte(t.Hour()))
	if err != nil {
		return "", err
	}

	var leftSb strings.Builder
	leftSb.WriteString(yearStr)
	leftSb.WriteByte(monthChar)
	leftSb.WriteByte(dayChar)
	leftSb.WriteByte(hourChar)

	minuteChar, err := encodeBase62(byte(t.Minute()))
	if err != nil {
		return "", err
	}

	secondChar, err := encodeBase62(byte(t.Second()))
	if err != nil {
		return "", err
	}

	var rightSb strings.Builder
	rightSb.WriteByte(minuteChar)
	rightSb.WriteByte(secondChar)

	var rightLen = format.SeparatorDepth + rightPrecision


	// if we need more right-hand digits, start subdividing the nanosecond
	// term into 1/60 fractionals
	if rightLen > 2 {
		var nanoSb strings.Builder
		var nanos float64 = float64(t.Nanosecond())
		var resolution float64 = 1_000_000_000
		for range rightLen - 2 {
			resolution = resolution / 60.0
			digit := int(nanos / resolution)
			digitChar, err := encodeBase62(uint8(digit))
			if err != nil {
				return "", err
			}
			nanoSb.WriteByte(digitChar)
			nanos -= float64(digit) * resolution
		}
		rightSb.WriteString(nanoSb.String())
	}

	var outSb strings.Builder
	leftStr := leftSb.String()
	rightStr := rightSb.String()
	sepIdx := format.SeparatorDepth
	outSb.WriteString(leftStr)
	outSb.WriteString(rightStr[:sepIdx])
	if rightPrecision > 0 {
		outSb.WriteByte(format.Separator)
		outSb.WriteString(rightStr[sepIdx:])
	}

	return outSb.String(), nil
}
