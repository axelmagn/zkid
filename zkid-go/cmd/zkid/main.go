package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/axelmagn/zkid/zkid-go"
)

func main() {
	var (
		rightPrecision uint
		minYearWidth   uint
		century        uint
		separator      string
		separatorDepth uint
	)

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

	encoded, err := zkid.Encode(time.Now(), format, uint8(minYearWidth), uint8(rightPrecision))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(encoded)
}
