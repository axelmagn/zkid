# ZKID

ZKID is a compact timestamp slug generator for human notetaking.

In notetaking systems such as zettelkasten, it is common practice to assign
unique identifiers to note files using timestamps. In some cases, the
timestamp is the entirety of the filename.

ZKID generates highly compact timestamps that remain human-readable. It aims to
balance compactness, ease of use, and good taste.

## Project Goals

- A simple, compact text timestamp format
- Lexicographically ordered
- Stable to epoch changes
- Separation of fields (a single character corresponds to exactly one field)
- Taste & aesthetics

## Format

Years are written in base 10. All other fields are written in base62.

ZKID strings encode wall-clock calendar and time components without timezone or
offset metadata; timezone is outside the format specification.

A typical compact timestamp looks like:

```
YYMDhm
268F6e (August 15th 2026, 06:40AM)
```

A typical full timestamp looks like:

```
YYYYMDhm-sffffff
20268F6e-8gVA0d4 (August 15th 2026, 06:40:08 AM)

YYYY:   year
M:      month
D:      day
h:      hour
m:      minute
s:      second
f:      subsecond fraction (1 / 60 of a second)
```

## Usage

### CLI

#### Installation

```bash
go install github.com/axelmagn/zkid/zkid-go/cmd/zkid@latest
```

#### Basic Usage

```bash
# Generate a compact timestamp for the current time
zkid
# Output: 2695K1

# Include seconds precision (1 digit right of separator)
zkid -p 1
# Output: 2695K1-8

# Use UTC timezone with subsecond precision
zkid -utc -p 5
# Output: 2695K1-8gVA0

# Encode timestamp(s) from standard input
echo "2026-08-15 06:40:00" | zkid
# Output: 268F6e

# Decode timestamp(s) from standard input (outputs datetime formatted string)
echo "268F6e" | zkid -decode
# Output: 2026-08-15 06:40:00
```

#### Flags

- `-p`, `-precision <n>`: Right precision digits (default: `0`).
- `-w`, `-year-width <n>`: Minimum year width/digits (default: `2`).
- `-utc`: Use UTC timezone instead of local timezone when generating timestamps.
- `-decode`: Decode ZKID timestamp slug(s) from standard input.
- `-c`, `-fmt-century <n>`: Format epoch century (default: `20`).
- `-s`, `-fmt-separator <char>`: Format separator character (default: `-`).
- `-d`, `-fmt-depth <n>`: Format separator depth (default: `4`).

### API

#### Go

```bash
go get github.com/axelmagn/zkid/zkid-go
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/axelmagn/zkid/zkid-go"
)

func main() {
	// Encode current time using default format (zkid 20-4)
	slug, err := zkid.Encode(time.Now(), zkid.DefaultFormat, 2, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println("Encoded:", slug)

	// Decode a timestamp slug (location is supplied by caller since ZKID does not encode timezone)
	t, err := zkid.Decode(slug, zkid.DefaultFormat, time.Local)
	if err != nil {
		panic(err)
	}
	fmt.Println("Decoded:", t.Format(time.RFC3339))
}
```

```go
type ZkidFormat struct {
	Century        uint16 // Epoch century (default: 20 for year 2000)
	Separator      byte   // Separator character (default: '-')
	SeparatorDepth uint8  // Separator depth (default: 4)
}

var DefaultFormat = ZkidFormat{
	Century:        20,
	Separator:      '-',
	SeparatorDepth: 4,
}

func Encode(t time.Time, format ZkidFormat, minYearWidth uint8, rightPrecision uint8) (string, error)
func Decode(s string, format ZkidFormat, loc *time.Location) (time.Time, error)
```

## Design

### Subformats

ZKID is a group of formats. The current default ZKID format is `zkid 20-4`.
ZKID formats are written:

```
CCsd

CC: epoch century
    19: year 1900
    20: year 2000
    21: year 2100
s: separator char
    - YYMDhm-sffffff
    . YYMDhm.sffffff
    : YYMDhm:sffffff
d: separator depth
    0: YY-MDhm...
    1: YYM-Dhm...
    2: YYMD-hm...
    3: YYMDh-m...
    4: YYMDhm-sffffff
    5: YYMDhms-ffffff
    6: YYMDhmsf-fffff
```

These variables determine the specifics of how ZKID strings are generated and
interpreted.

### Elide Millennium and Century by Default

These digits of the year can almost always be inferred from situational
context. If they cannot, it is easy enough to make them explicit retroactively.

### Elide Seconds and Milliseconds by Default

It is uncommon to need to know the exact second a note was created. The
primary purpose of seconds and fractional seconds is to deduplicate notes made
in the same minute.

### Separator and Separator Depth

The separator divides mandatory digits from non-mandatory right-hand digits. One
separator is sufficient to disambiguate extra year digits, and may be elided
almost all of the time.

### Base62 Digits and Field Ranges

We achieve a compact representation space by converting all digits except for
years to Base62 (`0-9`, `A-Z`, `a-z`).

| Field | Base | Valid Range |
| :--- | :--- | :--- |
| **Year** | 10 | `00`+ (variable width, minimum 2 digits matching epoch) |
| **Month** | 62 | `1`–`C` (1–12) |
| **Day** | 62 | `1`–`V` (1–31) |
| **Hour** | 62 | `0`–`N` (0–23) |
| **Minute** | 62 | `0`–`x` (0–59) |
| **Second** | 62 | `0`–`x` (0–59) |
| **Subsecond fraction** | 62 | `0`–`x` (0–59, each representing $\frac{1}{60^i}$ of a second) |

This encoding has several desirable properties:

- **Single-digit stability**: Each component occupies exactly one character.

- **ASCII sortability**: Base62 characters naturally follow ASCII ordering
(`0-9` < `A-Z` < `a-z`), ensuring timestamps sort lexicographically provided
year lengths are uniform.

- **Case distinction**: Month (1–12), Day (1–31), and Hour (0–23) values only
ever use `0-9` and uppercase letters `A-V`. Lowercase letters (`a-x`) only
appear in minute, second, and subsecond fraction positions.

### Subsecond Fractions in Base 60

After the hour digit, all subsequent digits (minute, second, subsecond
fractions) represent a 60th of the time unit that came before. Each fractional
digit $i$ (for $i \ge 1$) represents $\frac{1}{60^i}$ seconds. In almost all
practical cases, at most one subsecond digit is needed to break collisions.

### Epochs

The epoch year is determined by the epoch century of the ZKID format.

Years are digits written in base 10. All but the two lowest order digits may
be truncated if they match the epoch:

- If epoch is 2000 and the year is 2026, it is written `26`
- If epoch is 1900 and the year is 2026, it is written `2026`
- If epoch is 2000 and the year is 1990, it is written `1990`
- If epoch is 2000 and the year is 2142, it is written `142`

If for whatever reason the year needs to be extended beyond 9999, additional
year digits may be added ad infinitum.

### Decoding

When decoding, all elided year digits are assumed to match those of the epoch
year. All elided right-precision digits are assumed to be zero.

Because the ZKID format does not specify a timezone, the timezone of a decoded
timestamp is undefined by the specification. When decoding programmatically
(such as into a `time.Time` structure), the timezone/location must be explicitly
supplied by the consuming application or runtime context.

