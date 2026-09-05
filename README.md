# ZKID

ZKID is a compact timestamp slug generator for human notetaking.

In notetaking systems such as zettelkasten, it is common practice to assign
unique identifiers to note files using timestamps.  In some cases, the
timestamp is the entirety of the filename.

ZKID generates highly compact timestamp that remain human readable.  It aims to
balance compactness, ease of use, and good taste.

## Usage

```
zkid [options] [<datetime>]

TODO: options


```

## Project Goals

- A simple, compact text timestamp format
- Lexicographically ordered
- Stable to epoch changes
- Separation of fields (a single character corresponds to exactly one field)
- Taste & aesthetics.

## Usage

### CLI

TODO

### API

TODO

## Design

The full format looks like:

```
YYYYMD-HMS.FFFFFF
```

The minimal compaction looks like:

```
YYMD-H
```

### Observations

- Since compaction is possible on either side of the string, the hyphen separator
  is necessary to disambiguate between year digits and second / subsecond digits.
- A year string may be arbitrarily long, but by the time we get to the order of
  centuries, the digits are less and less useful because they are implicit to the
  historical context.  It will almost always be straightforward to infer the
  higher order digits of a year marker.
    - We can choose an "epoch" year (not necessarily 1970), which defines any
      higher order digits which are omitted.
    - If the year string is 23 and the epoch is 1900, the actual year is 1923.
    - If the year string is 023 and the epoch is 2100, the actual year is 2023.
    - the last two digits of the epoch are vestigial, but still included for
      clarity.
    - It is trivial to convert timestamps of one epoch to another, and only
      requires operating on the year characters.
- Once a timestamp is down to the hour resolution, every subsequent digit is a
  subdivision by 60 (60 minutes in an hour, 60 seconds in a minute).  This can be
  extended arbitrarily into subsecond divisions (60th of a second, 3600th of a
  second, etc).
    - a 60th of a second is a nice unit to work in.  It's roughly one frame of
      a well-optimized video game or desktop application.  It's small enough that
      it is highly unlikely that human input would be fast enough to lead to an
      ID collision.

### Base60 Digits

We achieve a compact representation space by converting all digits except for
years to base60 (truncated hex base62).  This encoding has a number of
desirable properties:

- numeric digits stay the same
- naturally follows ascii order (lexicographic sort is stable)
- intuitively simple
- covers full range of necessary values

### Years

Years are digits written in base 10.  All but the two lowest order digits may
be truncated if they match the epoch.

- If epoch is 2000 and the year is 2026, it is written 26
- If epoch is 1970 and the year is 2026, it is written 2026

If for whatever reason the year needs to be extended beyond 9999, additional
year digits may be added ad infinitum.

### Grammar

``` <zkid> ::= <year> <month> <day> <sep1> <hour> <minute> <second> <sep2>
<subsec>

<year> ::= <digit10> <digit10>+ <month> ::= <digit60> <day> ::= <digit60>

<hour> ::= <digit60> <minute> ::= <digit60> <second> ::= <digit60>

<subsec> ::= <digit60>+

```

