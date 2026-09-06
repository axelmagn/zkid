package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/axelmagn/zkid/zkid-go"
)

func main() {
	var (
		decode         bool
		utc            bool
		rightPrecision uint
		minYearWidth   uint
		century        uint
		separator      string
		separatorDepth uint
	)

	flag.BoolVar(&decode, "decode", false, "Decode ZKID timestamp slug(s) from standard input")
	flag.BoolVar(&utc, "utc", false, "Use UTC timezone instead of local")

	flag.UintVar(&rightPrecision, "precision", 0, "Right precision digits")
	flag.UintVar(&rightPrecision, "p", 0, "Right precision digits (shorthand)")

	flag.UintVar(&minYearWidth, "year-width", 2, "Minimum year width/digits")
	flag.UintVar(&minYearWidth, "w", 2, "Minimum year width/digits (shorthand)")

	flag.UintVar(&century, "fmt-century", uint(zkid.DefaultFormat.Century), "Format epoch century")
	flag.UintVar(&century, "c", uint(zkid.DefaultFormat.Century), "Format epoch century (shorthand)")

	flag.StringVar(&separator, "fmt-separator", string(zkid.DefaultFormat.Separator), "Format separator character")
	flag.StringVar(&separator, "s", string(zkid.DefaultFormat.Separator), "Format separator character (shorthand)")

	flag.UintVar(&separatorDepth, "fmt-depth", uint(zkid.DefaultFormat.SeparatorDepth), "Format separator depth")
	flag.UintVar(&separatorDepth, "d", uint(zkid.DefaultFormat.SeparatorDepth), "Format separator depth (shorthand)")

	flag.Parse()

	if len(separator) != 1 {
		fmt.Fprintf(os.Stderr, "error: separator must be a single character\n")
		os.Exit(1)
	}

	format := zkid.ZkidFormat{
		Century:        uint16(century),
		Separator:      separator[0],
		SeparatorDepth: uint8(separatorDepth),
	}

	loc := time.Local
	if utc {
		loc = time.UTC
	}

	if decode {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			decoded, err := zkid.Decode(line, format, loc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error decoding %q: %v\n", line, err)
				os.Exit(1)
			}
			fmt.Println(decoded.Format(time.DateTime))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading standard input: %v\n", err)
			os.Exit(1)
		}
		return
	}

	stat, err := os.Stdin.Stat()
	hasStdin := err == nil && (stat.Mode()&os.ModeCharDevice) == 0

	if hasStdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			t, err := time.ParseInLocation(time.DateTime, line, loc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing datetime %q: %v\n", line, err)
				os.Exit(1)
			}
			encoded, err := zkid.Encode(t, format, uint8(minYearWidth), uint8(rightPrecision))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error encoding %q: %v\n", line, err)
				os.Exit(1)
			}
			fmt.Println(encoded)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading standard input: %v\n", err)
			os.Exit(1)
		}
		return
	}

	now := time.Now().In(loc)
	encoded, err := zkid.Encode(now, format, uint8(minYearWidth), uint8(rightPrecision))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(encoded)
}
