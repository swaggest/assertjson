package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/swaggest/assertjson"
)

func main() {
	var (
		input, output  string
		length         int
		prefix, indent string
	)

	flag.StringVar(&output, "output", "", "path to output json file, if not specified input file is used")
	flag.IntVar(&length, "len", 100, "line length limit")
	flag.StringVar(&prefix, "prefix", "", "prefix")
	flag.StringVar(&indent, "indent", " ", "indent")
	flag.Parse()

	input = flag.Arg(0)
	if input == "" {
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Missing input path argument.")
		flag.Usage()

		return
	}

	if output == "" {
		output = input
	}

	// nolint:gosec // Intentional file reading.
	data, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatalf("could not read input: %v", err)
	}

	data, err = assertjson.MarshalIndentCompact(json.RawMessage(data), "", "  ", length)
	if err != nil {
		log.Fatalf("could not process input: %v", err)
	}

	err = ioutil.WriteFile(output, data, 0o600)
	if err != nil {
		log.Fatalf("could not write output: %v", err)
	}
}
