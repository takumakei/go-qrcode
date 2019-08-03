// go-qrcode
// Copyright 2014 Tom Harwood

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	outFile := flag.String("o", "", "out PNG file prefix, empty for stdout")
	size := flag.Int("s", 256, "image size (pixel)")
	textArt := flag.Bool("t", false, "print as text-art on stdout")
	negative := flag.Bool("i", false, "invert black and white")
	page := flag.String("p", "", "structured append mode. 'current/last:parity' e.g. '2/3:0x11'\ncurrent and last must be 1 to 16.\nparity can be decimal or hex (with 0x prefix), 0 to 255.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `qrcode -- QR Code encoder in Go
https://github.com/skip2/go-qrcode

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Usage:
  1. Arguments except for flags are joined by " " and used to generate QR code.
     Default output is STDOUT, pipe to imagemagick command "display" to display
     on any X server.

       qrcode hello word | display

  2. Save to file if "display" not available:

       qrcode "homepage: https://github.com/skip2/go-qrcode" > out.png

`)
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		checkError(fmt.Errorf("Error: no content given"))
	}

	content := strings.Join(flag.Args(), " ")

	var err error
	var q *qrcode.QRCode
	q, err = qrcode.NewWithPage(content, qrcode.Highest, parsePage(*page))
	checkError(err)

	if *textArt {
		art := q.ToString(*negative)
		fmt.Println(art)
		return
	}

	if *negative {
		q.ForegroundColor, q.BackgroundColor = q.BackgroundColor, q.ForegroundColor
	}

	var png []byte
	png, err = q.PNG(*size)
	checkError(err)

	if *outFile == "" {
		os.Stdout.Write(png)
	} else {
		var fh *os.File
		fh, err = os.Create(*outFile + ".png")
		checkError(err)
		defer fh.Close()
		fh.Write(png)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func parsePage(arg string) *qrcode.Page {
	if len(arg) == 0 {
		return nil
	}

	m := regexp.MustCompile(`^([1-9][0-9]?)/([1-9][0-9]?):(?:0x([0-9a-fA-F]{1,2})|([1-9][0-9]{0,2}|0))$`).FindStringSubmatch(arg)
	if len(m) != 5 {
		invalidPage(arg)
	}

	current, err := strconv.Atoi(m[1])
	if err != nil || current < 1 || 16 < current {
		invalidPage(arg)
	}

	last, err := strconv.Atoi(m[2])
	if err != nil || last < 1 || 16 < last {
		invalidPage(arg)
	}

	var parity byte
	switch {
	case len(m[3]) != 0:
		if v, err := strconv.ParseInt(m[3], 16, 64); err != nil || v < 0 || v > 255 {
			invalidPage(arg)
		} else {
			parity = byte(v)
		}
	case len(m[4]) != 0:
		if v, err := strconv.ParseInt(m[4], 10, 64); err != nil || v < 0 || v > 255 {
			invalidPage(arg)
		} else {
			parity = byte(v)
		}
	default:
		invalidPage(arg)
	}

	return &qrcode.Page{byte(current - 1), byte(last - 1), parity}
}

func invalidPage(arg string) {
	log.Panicf("invalid argument '-a' '%s'", arg)
}
