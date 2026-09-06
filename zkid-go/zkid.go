package zkid

import (
	"errors"
	"fmt"
	// "math"
	"strings"
	"time"
)

var (
	ErrInvalidFormat    = errors.New("invalid zkid format")
	ErrInvalidTimestamp = errors.New("invalid zkid timestamp")
	ErrInvalidSeparator = errors.New("invalid separator character")
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
	SeparatorDepth uint8  // Separator depth (from year term)
}

// DefaultFormat is the default ZKID format (zkid 20-4).
var DefaultFormat = ZkidFormat{
	Century:        20,
	Separator:      '-',
	SeparatorDepth: 4,
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

func extractBase62(s string, idx int) (byte, error) {
	if idx >= len(s) {
		return 0, errors.ErrUnsupported
	}
	var fieldStr string
	scanned, err := fmt.Sscanf(s[idx:idx+1], "%s", &fieldStr)
	if err != nil {
		return 0, err
	}
	if scanned != 1 || len(fieldStr) != 1 {
		panic("assertion failed: did not parse expected number of fields")
	}
	return decodeBase62(fieldStr[0])
}

// Encode encodes a time.Time into a ZKID string using the given format.
func Encode(t time.Time, format ZkidFormat, minYearWidth uint8, rightPrecision uint8) (string, error) {
	year := t.Year()
	if year < 0 ||
		t.Month() < 1 || t.Month() > 12 ||
		t.Day() < 1 || t.Day() > 31 ||
		t.Hour() < 0 || t.Hour() > 23 ||
		t.Minute() < 0 || t.Minute() > 59 ||
		t.Second() < 0 || t.Second() > 59 ||
		t.Nanosecond() < 0 || t.Nanosecond() >= 1_000_000_000 {
		return "", ErrDateOutOfRange
	}

	if format.Separator < 33 || format.Separator > 126 || decodeBase62Table[format.Separator] != -1 {
		return "", ErrInvalidSeparator
	}

	if uint16(format.SeparatorDepth)+uint16(rightPrecision) > 255 {
		return "", ErrInvalidFormat
	}

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

	minuteChar, err := encodeBase62(byte(t.Minute()))
	if err != nil {
		return "", err
	}

	secondChar, err := encodeBase62(byte(t.Second()))
	if err != nil {
		return "", err
	}

	var rightSb strings.Builder
	rightSb.WriteByte(monthChar)
	rightSb.WriteByte(dayChar)
	rightSb.WriteByte(hourChar)
	rightSb.WriteByte(minuteChar)
	rightSb.WriteByte(secondChar)

	totalLen := int(format.SeparatorDepth) + int(rightPrecision)
	rightLen := rightSb.Len()
	if totalLen > rightLen {
		var nanos float64 = float64(t.Nanosecond())
		var resolution float64 = 1_000_000_000
		for range totalLen - rightLen {
			resolution = resolution / 60.0
			digit := int(nanos / resolution)
			digitChar, err := encodeBase62(uint8(digit))
			if err != nil {
				return "", err
			}
			rightSb.WriteByte(digitChar)
			nanos -= float64(digit) * resolution
		}
	}

	rightStr := rightSb.String()
	if len(rightStr) > totalLen {
		rightStr = rightStr[:totalLen]
	}

	var outSb strings.Builder
	outSb.WriteString(yearStr)
	sepIdx := int(format.SeparatorDepth)
	outSb.WriteString(rightStr[:sepIdx])
	if rightPrecision > 0 {
		outSb.WriteByte(format.Separator)
		outSb.WriteString(rightStr[sepIdx:])
	}

	return outSb.String(), nil
}

// Decode decodes a ZKID string into a time.Time using the specified format and location.
func Decode(s string, format ZkidFormat, loc *time.Location) (time.Time, error) {
	if loc == nil {
		return time.Time{}, ErrInvalidFormat
	}

	if format.Separator < 33 || format.Separator > 126 || decodeBase62Table[format.Separator] != -1 {
		return time.Time{}, ErrInvalidSeparator
	}

	sLen := len(s)
	sepIdx := sLen
	for i := range sLen {
		if s[i] == format.Separator {
			sepIdx = i
			break
		}
	}

	monthIdx := sepIdx - int(format.SeparatorDepth)
	if monthIdx < 1 {
		return time.Time{}, ErrInvalidTimestamp
	}

	var year uint16
	yearStr := s[:monthIdx]
	scanned, err := fmt.Sscanf(yearStr, "%d", &year)
	if err != nil || scanned != 1 {
		return time.Time{}, ErrInvalidTimestamp
	}
	// add any elided epoch digits
	epoch := format.Century * 100
	unelided := year
	var epochFactor uint16 = 1
	for epoch > 0 && unelided > 0 {
		epoch /= 10
		unelided /= 10
		epochFactor *= 10
	}
	year += epoch * epochFactor



	var fields [5]uint8 // ASSUME: zero initialized
	fields[0] = 1
	fields[1] = 1
	var fieldIdx = 0
	var sIdx = monthIdx
	fieldsLen := len(fields)
	// scan fields to left of separator
	for fieldIdx < fieldsLen && sIdx < sepIdx {
		field, err := extractBase62(s, sIdx)
		if err != nil {
			return time.Time{}, err
		}
		fields[fieldIdx] = field
		fieldIdx++
		sIdx++
	}

	// scan fields to right of separator
	sIdx++
	for fieldIdx < fieldsLen && sIdx < sLen {
		field, err := extractBase62(s, sIdx)
		if err != nil {
			return time.Time{}, err
		}

		fields[fieldIdx] = field
		fieldIdx++
		sIdx++
	}
	month := fields[0]
	day := fields[1]
	hour := fields[2]
	minute := fields[3]
	second := fields[4]

	// scan nanoseconds
	var nanos float64 = 0
	var resolution float64 = 1_000_000_000
	for sIdx < sLen {
		resolution /= 60.0
		field, err := extractBase62(s, sIdx)
		if err != nil {
			return time.Time{}, err
		}
		nanos += float64(field) * resolution
		sIdx++
	}

	if month < 1 || month > 12 ||
		day < 1 || day > 31 ||
		hour > 23 ||
		minute > 59 ||
		second > 59 ||
		nanos < 0 || nanos >= 1_000_000_000 {
		return time.Time{}, ErrDateOutOfRange
	}

	t := time.Date(
		int(year),
		time.Month(month),
		int(day),
		int(hour),
		int(minute),
		int(second),
		int(nanos),
		loc)

	// check for leap year shenanigans
	if t.Month() != time.Month(month) || t.Day() != int(day) {
		return time.Time{}, ErrDateOutOfRange
	}

	return t, nil
}
