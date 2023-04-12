// Package main provides a tool to compact JSON.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/bool64/dev/version"
	"github.com/swaggest/assertjson"
)

// The function is a bit lengthy, but I'm not sure if it would be more approachable divided in several functions.
func main() { //nolint
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
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Missing input path argument, use `-` for stdin.")
		flag.Usage()

		return
	}

	// Read stdin.
	if input == "-" {
		var v interface{}

		dec := json.NewDecoder(os.Stdin)

		err := dec.Decode(&v)
		if err != nil {
			log.Fatalf("could not process input: %v", err)
		}

		comp, err := assertjson.MarshalIndentCompact(v, prefix, indent, length)
		if err != nil {
			log.Fatalf("could not process input: %v", err)
		}

		fmt.Println(string(comp))

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

			//nolint:gosec // Intentional file reading.
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

			err = ioutil.WriteFile(out, comp, 0o600)
			if err != nil {
				log.Fatalf("could not write output to %s: %v", out, err)
			}
		}
	}
}
