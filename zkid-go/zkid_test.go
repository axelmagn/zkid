package zkid

import (
	"errors"
	"testing"
	"time"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	loc := time.UTC
	testCases := []struct {
		name           string
		t              time.Time
		format         ZkidFormat
		minYearWidth   uint8
		rightPrecision uint8
	}{
		{
			name:           "default format 20-4 basic",
			t:              time.Date(2026, 9, 5, 18, 7, 0, 0, loc),
			format:         DefaultFormat,
			minYearWidth:   2,
			rightPrecision: 0,
		},
		{
			name:           "default format 20-4 with seconds",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         DefaultFormat,
			minYearWidth:   2,
			rightPrecision: 1,
		},
		{
			name:           "full year width 4 digits",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         DefaultFormat,
			minYearWidth:   4,
			rightPrecision: 1,
		},
		{
			name:           "1900 century format",
			t:              time.Date(1990, 8, 15, 6, 40, 8, 0, loc),
			format:         ZkidFormat{Century: 19, Separator: '-', SeparatorDepth: 4},
			minYearWidth:   2,
			rightPrecision: 1,
		},
		{
			name:           "1900 timestamp encoded with 2000 century (explicit 4-digit year)",
			t:              time.Date(1990, 8, 15, 6, 40, 8, 0, loc),
			format:         DefaultFormat,
			minYearWidth:   2,
			rightPrecision: 1,
		},
		{
			name:           "subsecond precision 1 fraction digit",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 500_000_000, loc),
			format:         DefaultFormat,
			minYearWidth:   2,
			rightPrecision: 2, // minute is depth 4, so totalLen=6 (minute, second, frac1)
		},
		{
			name:           "separator depth 0 (year before separator)",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 0},
			minYearWidth:   2,
			rightPrecision: 5,
		},
		{
			name:           "separator depth 1 (month before separator)",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 1},
			minYearWidth:   2,
			rightPrecision: 4,
		},
		{
			name:           "separator depth 2 (day before separator)",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         ZkidFormat{Century: 20, Separator: '.', SeparatorDepth: 2},
			minYearWidth:   2,
			rightPrecision: 3,
		},
		{
			name:           "separator depth 3 (hour before separator)",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 3},
			minYearWidth:   2,
			rightPrecision: 2,
		},
		{
			name:           "separator depth 5 (seconds before separator)",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, loc),
			format:         ZkidFormat{Century: 20, Separator: '.', SeparatorDepth: 5},
			minYearWidth:   2,
			rightPrecision: 0,
		},
		{
			name:           "separator depth 5 with subseconds",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 500_000_000, loc),
			format:         ZkidFormat{Century: 20, Separator: ':', SeparatorDepth: 5},
			minYearWidth:   2,
			rightPrecision: 1,
		},
		{
			name:           "non-UTC timezone",
			t:              time.Date(2026, 9, 5, 18, 7, 30, 0, time.FixedZone("PST", -8*3600)),
			format:         DefaultFormat,
			minYearWidth:   2,
			rightPrecision: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := Encode(tc.t, tc.format, tc.minYearWidth, tc.rightPrecision)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(encoded, tc.format, tc.t.Location())
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !decoded.Equal(tc.t) {
				t.Errorf("Decoded time %v does not match original %v (encoded string: %s)", decoded, tc.t, encoded)
			}
		})
	}
}

func TestReadmeExamples(t *testing.T) {
	// README example: YYMDhm 908F6e (August 15th 1990, 06:40AM)
	// Century: 19
	fmt19 := ZkidFormat{Century: 19, Separator: '-', SeparatorDepth: 4}
	d, err := Decode("908F6e", fmt19, time.UTC)
	if err != nil {
		t.Fatalf("Failed to decode readme example 1: %v", err)
	}
	expected := time.Date(1990, time.August, 15, 6, 40, 0, 0, time.UTC)
	if !d.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, d)
	}
}

func TestDecodeOmittedFields(t *testing.T) {
	loc := time.UTC
	testCases := []struct {
		name     string
		input    string
		format   ZkidFormat
		expected time.Time
	}{
		{
			name:     "depth 0 year only",
			input:    "26",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 0},
			expected: time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
		},
		{
			name:     "depth 1 year and month",
			input:    "268",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 1},
			expected: time.Date(2026, 8, 1, 0, 0, 0, 0, loc),
		},
		{
			name:     "depth 2 year month day",
			input:    "268F",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 2},
			expected: time.Date(2026, 8, 15, 0, 0, 0, 0, loc),
		},
		{
			name:     "depth 2 year month day with separated hour minute",
			input:    "268F-6e",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 2},
			expected: time.Date(2026, 8, 15, 6, 40, 0, 0, loc),
		},
		{
			name:     "depth 3 year month day hour",
			input:    "268F6",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 3},
			expected: time.Date(2026, 8, 15, 6, 0, 0, 0, loc),
		},
		{
			name:     "depth 3 year month day hour trailing separator",
			input:    "268F6-",
			format:   ZkidFormat{Century: 20, Separator: '-', SeparatorDepth: 3},
			expected: time.Date(2026, 8, 15, 6, 0, 0, 0, loc),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := Decode(tc.input, tc.format, loc)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}
			if !decoded.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, decoded)
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	format := DefaultFormat

	testCases := []struct {
		name        string
		input       string
		format      ZkidFormat
		loc         *time.Location
		expectedErr error
	}{
		{
			name:        "nil location",
			input:       "2695I7",
			format:      format,
			loc:         nil,
			expectedErr: ErrInvalidFormat,
		},
		{
			name:        "invalid separator in format",
			input:       "2695I7",
			format:      ZkidFormat{Century: 20, Separator: '0', SeparatorDepth: 4},
			loc:         time.UTC,
			expectedErr: ErrInvalidSeparator,
		},
		{
			name:        "multiple separators",
			input:       "2695I7-1-2",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrInvalidTimestamp,
		},
		{
			name:        "string too short",
			input:       "269",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrInvalidTimestamp,
		},
		{
			name:        "non-decimal year",
			input:       "AB95I7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrInvalidTimestamp,
		},
		{
			name:        "month out of range (0)",
			input:       "2605I7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "month out of range (>12, 'D'=13)",
			input:       "26D5I7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "day out of range (0)",
			input:       "2690I7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "day out of range (>31, 'W'=32)",
			input:       "269WI7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "hour out of range (>23, 'O'=24)",
			input:       "2695O7",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "minute out of range (>=60, 'y'=60)",
			input:       "2695Iy",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "second out of range (>=60, 'z'=61)",
			input:       "2695I7-z",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "fraction out of range (>=60)",
			input:       "2695I7-0y",
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
		{
			name:        "invalid calendar date (Feb 30)",
			input:       "262UI7", // Month=2 (Feb), Day='U' (30)
			format:      format,
			loc:         time.UTC,
			expectedErr: ErrDateOutOfRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decode(tc.input, tc.format, tc.loc)
			if err == nil {
				t.Fatalf("Expected error %v, got nil", tc.expectedErr)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
