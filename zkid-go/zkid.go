package zkid

import (
	"errors"
	"fmt"
	"math"
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

// Decode decodes a ZKID string into a time.Time using the specified format and location.
func Decode(s string, format ZkidFormat, loc *time.Location) (time.Time, error) {
	if loc == nil {
		return time.Time{}, ErrInvalidFormat
	}

	if format.Separator < 33 || format.Separator > 126 || decodeBase62Table[format.Separator] != -1 {
		return time.Time{}, ErrInvalidSeparator
	}

	if format.SeparatorDepth == 0 {
		return time.Time{}, ErrInvalidFormat
	}

	var leftStr, rightSuffix string
	sepCount := strings.Count(s, string(format.Separator))
	if sepCount > 1 {
		return time.Time{}, ErrInvalidTimestamp
	} else if sepCount == 1 {
		parts := strings.Split(s, string(format.Separator))
		leftStr = parts[0]
		rightSuffix = parts[1]
		if len(rightSuffix) == 0 {
			return time.Time{}, ErrInvalidTimestamp
		}
	} else {
		leftStr = s
		rightSuffix = ""
	}

	minLeftLen := 1 + 3 + int(format.SeparatorDepth)
	if len(leftStr) < minLeftLen {
		return time.Time{}, ErrInvalidTimestamp
	}

	yearStrEnd := len(leftStr) - 3 - int(format.SeparatorDepth)
	yearStr := leftStr[:yearStrEnd]
	monthChar := leftStr[yearStrEnd]
	dayChar := leftStr[yearStrEnd+1]
	hourChar := leftStr[yearStrEnd+2]
	rightPrefix := leftStr[yearStrEnd+3:]

	for i := 0; i < len(yearStr); i++ {
		if yearStr[i] < '0' || yearStr[i] > '9' {
			return time.Time{}, ErrInvalidTimestamp
		}
	}

	epoch := format.Century * 100
	epochStr := fmt.Sprintf("%d", epoch)
	var fullYearStr string
	if len(yearStr) < len(epochStr) {
		fullYearStr = epochStr[:len(epochStr)-len(yearStr)] + yearStr
	} else {
		fullYearStr = yearStr
	}

	var year int
	_, err := fmt.Sscanf(fullYearStr, "%d", &year)
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}

	monthVal, err := decodeBase62(monthChar)
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}
	if monthVal < 1 || monthVal > 12 {
		return time.Time{}, ErrDateOutOfRange
	}

	dayVal, err := decodeBase62(dayChar)
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}
	if dayVal < 1 || dayVal > 31 {
		return time.Time{}, ErrDateOutOfRange
	}

	hourVal, err := decodeBase62(hourChar)
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}
	if hourVal > 23 {
		return time.Time{}, ErrDateOutOfRange
	}

	fullRightStr := rightPrefix + rightSuffix
	if len(fullRightStr) == 0 {
		return time.Time{}, ErrInvalidTimestamp
	}

	minVal, err := decodeBase62(fullRightStr[0])
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}
	if minVal > 59 {
		return time.Time{}, ErrDateOutOfRange
	}

	var secVal byte = 0
	if len(fullRightStr) > 1 {
		secVal, err = decodeBase62(fullRightStr[1])
		if err != nil {
			return time.Time{}, ErrInvalidTimestamp
		}
		if secVal > 59 {
			return time.Time{}, ErrDateOutOfRange
		}
	}

	var nsec int = 0
	if len(fullRightStr) > 2 {
		var nanos float64 = 0
		var resolution float64 = 1_000_000_000
		for i := 2; i < len(fullRightStr); i++ {
			resolution /= 60.0
			digitVal, err := decodeBase62(fullRightStr[i])
			if err != nil {
				return time.Time{}, ErrInvalidTimestamp
			}
			if digitVal > 59 {
				return time.Time{}, ErrDateOutOfRange
			}
			nanos += float64(digitVal) * resolution
		}
		nsec = int(math.Round(nanos))
		if nsec >= 1_000_000_000 {
			nsec = 999_999_999
		}
	}

	t := time.Date(year, time.Month(monthVal), int(dayVal), int(hourVal), int(minVal), int(secVal), nsec, loc)
	if t.Month() != time.Month(monthVal) || t.Day() != int(dayVal) {
		return time.Time{}, ErrDateOutOfRange
	}

	return t, nil
}
