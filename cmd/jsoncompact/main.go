package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/bool64/dev/version"
	"github.com/swaggest/assertjson"
)

// nolint // The function is a bit lengthy, but I'm not sure it
// would be more approachable if split in several functions.
func main() {
	var (
		input, output  string
		length         int
		prefix, indent string
		ver, verbose   bool
	)

	flag.StringVar(&output, "output", "", "Path to output json file, if not specified input file is used.")
	flag.IntVar(&length, "len", 100, "Line length limit.")
	flag.StringVar(&prefix, "prefix", "", "Set prefix.")
	flag.StringVar(&indent, "indent", " ", "Set indent.")
	flag.BoolVar(&ver, "version", false, "Print version and exit.")
	flag.BoolVar(&verbose, "v", false, "Verbose mode.")
	flag.Parse()

	if ver {
		fmt.Println(version.Info().Version)

		return
	}

	input = flag.Arg(0)
	if input == "" {
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Missing input path argument.")
		flag.Usage()

		return
	}

	for _, in := range flag.Args() {
		matches, err := filepath.Glob(in)
		if err != nil {
			log.Fatalf("could not read input: %v", err)
		}

		for _, m := range matches {
			if verbose {
				log.Printf("compacting %s.\n", m)
			}

			// nolint:gosec // Intentional file reading.
			orig, err := ioutil.ReadFile(m)
			if err != nil {
				log.Fatalf("could not read input %s: %v", m, err)
			}

			comp, err := assertjson.MarshalIndentCompact(json.RawMessage(orig), prefix, indent, length)
			if err != nil {
				log.Fatalf("could not process input: %v", err)
			}

			if bytes.Equal(orig, comp) {
				if verbose {
					log.Printf("already compact, skipping %s\n", m)
				}

				continue
			}

			out := output
			if out == "" {
				out = m
			}

			if verbose {
				log.Printf("writing to %s\n", out)
			}

			err = ioutil.WriteFile(out, comp, 0600)
			if err != nil {
				log.Fatalf("could not write output to %s: %v", out, err)
			}
		}
	}
}
