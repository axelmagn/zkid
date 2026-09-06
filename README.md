# ZKID

ZKID is a compact timestamp slug generator for human notetaking.

In notetaking systems such as zettelkasten, it is common practice to assign
unique identifiers to note files using timestamps.  In some cases, the
timestamp is the entirety of the filename.

ZKID generates highly compact timestamp that remain human readable.  It aims to
balance compactness, ease of use, and good taste.

## Format

Years are written in base 10. All other digits are written in base62.

A typical compact timestamp looks like:

```
YYMDhm
908F6e (August 15th 1990, 06:40AM)
```

A typical full timestamp looks like:

```
YYYYMDhm-sffffff
19908F6e-8gVA0d4 (August 15th 1990, 06:40AM)

YYYY:   year
M:      month
D:      day
h:      hour
m:      minute
s:      second
f:      subsecond fraction (1 / 60)
```

## Usage

### CLI

TODO

### API

#### Go

```go

type ZkidFormat struct {
    Century uint16
    Separator rune
    SeparatorDepth uint8
}

func Encode(time Time, format ZkidFormat, minYearDigits uint8, rightPrecision uint8) (string, error)
func Decode(string, format ZkidFormat) (Time, error)

```

## Project Goals

- A simple, compact text timestamp format
- Lexicographically ordered
- Stable to epoch changes
- Separation of fields (a single character corresponds to exactly one field)
- Taste & aesthetics.

## Design

### subformats

zkid is a group of formats.  The current default zkid format is `zkid 20-1`.
zkid formats are written:

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
    1: YYMDhm.sffffff
    2: YYMDhms.ffffff
    3: YYMDhmsf.fffff
```

These variables determine the specifics of how zkid strings are generated and
interpreted.

### elide millenium and century by default

These digits of the year can almost always be inferred from situational
context.  If the cannot, is easy enough to make them explicit retroactively.

### elide seconds and milliseconds by default

It is uncommon to need to know the exact second a note was created.  The
primary purpose of seconds and fractional seconds is to deduplicate notes made
in the same minute.

### second to the right of separator

Separator divides mandatory digits from non-mandatory right digits.  One
separator is sufficient to disambiguate extra year digits, and may be elided
almost all of the time.

### Base62 Digits

We achieve a compact representation space by converting all digits except for
years to base62.  This encoding has a number of
desirable properties:

- single digit numbers are stable
- naturally follows ascii order (lexicographic sort is stable)
- intuitively simple
- covers full range of necessary values
- datestamps and hours map naturally to hex32 (case independent)

### Subsecond fractions in base 60

After the hour digit, all subsequent digits (minute, second, subsecond)
represent a 60th of the time unit that came before.  Extending this into
subseconds is intuitive, and in almost all cases only the first subsecond digit
would ever be necessary.


### epochs

The epoch year is determined by the epoch century of the zkid format.

Years are digits written in base 10.  All but the two lowest order digits may
be truncated if they match the epoch.

- If epoch is 2000 and the year is 2026, it is written 26
- If epoch is 1900 and the year is 2026, it is written 2026

If for whatever reason the year needs to be extended beyond 9999, additional
year digits may be added ad infinitum.

### decoding

when decoding, all elided year digits are assumed to match those of the epoch
year.  all elided right-precision bits are assumed to be zero.

