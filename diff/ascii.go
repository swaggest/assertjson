package diff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// NewASCIIFormatter creates a new ASCIIFormatter instance with the specified left data and configuration settings.
func NewASCIIFormatter(left interface{}, config ASCIIFormatterConfig) *ASCIIFormatter {
	return &ASCIIFormatter{
		left:   left,
		config: config,
	}
}

// ASCIIFormatter is used to generate ASCII representations of differences between JSON-like data structures.
type ASCIIFormatter struct {
	left    interface{}
	config  ASCIIFormatterConfig
	buffer  *bytes.Buffer
	path    []string
	size    []int
	inArray []bool
	line    *ASCIILine
}

// ASCIIFormatterConfig specifies configuration options for formatting ASCII representations of data structures.
// ShowArrayIndex determines if array indices should be displayed in the formatted output.
// Coloring enables or disables colored output in the formatted result.
type ASCIIFormatterConfig struct {
	ShowArrayIndex bool
	Coloring       bool
}

// ASCIILine represents a line in an ASCII-formatted output with a marker, indentation, and a buffer containing content.
type ASCIILine struct {
	marker string
	indent int
	buffer *bytes.Buffer
}

// Format formats the differences between two JSON objects into an ASCII representation using the provided Diff.
func (f *ASCIIFormatter) Format(diff Diff) (result string, err error) {
	f.buffer = bytes.NewBuffer([]byte{})
	f.path = []string{}
	f.size = []int{}
	f.inArray = []bool{}

	if v, ok := f.left.(map[string]interface{}); ok {
		f.formatObject(v, diff)
	} else if v, ok := f.left.([]interface{}); ok {
		f.formatArray(v, diff)
	} else {
		return "", fmt.Errorf("expected map[string]interface{} or []interface{}, got %T",
			f.left)
	}

	return f.buffer.String(), nil
}

func (f *ASCIIFormatter) formatObject(left map[string]interface{}, df Diff) {
	f.addLineWith("{")
	f.push("ROOT", len(left), false)
	f.processObject(left, df.Deltas())
	f.pop()
	f.addLineWith("}")
}

func (f *ASCIIFormatter) formatArray(left []interface{}, df Diff) {
	f.addLineWith("[")
	f.push("ROOT", len(left), true)
	f.processArray(left, df.Deltas())
	f.pop()
	f.addLineWith("]")
}

func (f *ASCIIFormatter) processArray(array []interface{}, deltas []Delta) error {
	patchedIndex := 0

	for index, value := range array {
		if err := f.processItem(value, deltas, Index(index)); err != nil {
			return err
		}

		patchedIndex++
	}

	// additional Added
	for _, delta := range deltas {
		switch delta.(type) {
		case *Added:
			d := delta.(*Added)
			// skip items already processed
			if int(d.Position.(Index)) < len(array) {
				continue
			}

			f.printRecursive(d.String(), d.Value, ASCIIAdded)
		}
	}

	return nil
}

func (f *ASCIIFormatter) processObject(object map[string]interface{}, deltas []Delta) error {
	names := sortedKeys(object)
	for _, name := range names {
		value := object[name]
		if err := f.processItem(value, deltas, Name(name)); err != nil {
			return err
		}
	}

	// Added
	for _, delta := range deltas {
		switch dt := delta.(type) {
		case *Added:
			d := dt
			f.printRecursive(d.String(), d.Value, ASCIIAdded)
		}
	}

	return nil
}

func (f *ASCIIFormatter) processItem(value interface{}, deltas []Delta, position Position) error {
	matchedDeltas := f.searchDeltas(deltas, position)
	positionStr := position.String()

	if len(matchedDeltas) > 0 {
		for _, matchedDelta := range matchedDeltas {
			switch matchedDelta.(type) {
			case *Object:
				d := matchedDelta.(*Object)

				switch value.(type) {
				case map[string]interface{}:
					// ok
				default:
					return errors.New("Type mismatch")
				}

				o := value.(map[string]interface{})

				f.newLine(ASCIISame)
				f.printKey(positionStr)
				f.print("{")
				f.closeLine()
				f.push(positionStr, len(o), false)
				f.processObject(o, d.Deltas)
				f.pop()
				f.newLine(ASCIISame)
				f.print("}")
				f.printComma()
				f.closeLine()

			case *Array:
				d := matchedDelta.(*Array)

				switch value.(type) {
				case []interface{}:
					// ok
				default:
					return errors.New("Type mismatch")
				}

				a := value.([]interface{})

				f.newLine(ASCIISame)
				f.printKey(positionStr)
				f.print("[")
				f.closeLine()
				f.push(positionStr, len(a), true)
				f.processArray(a, d.Deltas)
				f.pop()
				f.newLine(ASCIISame)
				f.print("]")
				f.printComma()
				f.closeLine()

			case *Added:
				d := matchedDelta.(*Added)
				f.printRecursive(positionStr, d.Value, ASCIIAdded)

				f.size[len(f.size)-1]++

			case *Modified:
				d := matchedDelta.(*Modified)
				savedSize := f.size[len(f.size)-1]
				f.printRecursive(positionStr, d.OldValue, ASCIIDeleted)
				f.size[len(f.size)-1] = savedSize
				f.printRecursive(positionStr, d.NewValue, ASCIIAdded)

			case *TextDiff:
				savedSize := f.size[len(f.size)-1]
				d := matchedDelta.(*TextDiff)
				f.printRecursive(positionStr, d.OldValue, ASCIIDeleted)
				f.size[len(f.size)-1] = savedSize
				f.printRecursive(positionStr, d.NewValue, ASCIIAdded)

			case *Deleted:
				d := matchedDelta.(*Deleted)
				f.printRecursive(positionStr, d.Value, ASCIIDeleted)

			default:
				return errors.New("Unknown Delta type detected")
			}
		}
	} else {
		f.printRecursive(positionStr, value, ASCIISame)
	}

	return nil
}

func (f *ASCIIFormatter) searchDeltas(deltas []Delta, position Position) (results []Delta) {
	results = make([]Delta, 0)

	for _, delta := range deltas {
		switch dt := delta.(type) {
		case PostDelta:
			if dt.PostPosition() == position {
				results = append(results, delta)
			}
		case PreDelta:
			if dt.PrePosition() == position {
				results = append(results, delta)
			}
		default:
			panic("heh")
		}
	}

	return results
}

const (
	// ASCIISame represents the ASCII string " " used to indicate unchanged or identical elements in a comparison.
	ASCIISame = " "

	// ASCIIAdded represents the ASCII string "+" used to indicate newly added items or elements in comparisons.
	ASCIIAdded = "+"

	// ASCIIDeleted represents the ASCII string "-" used to indicate deleted items or removed elements.
	ASCIIDeleted = "-"
)

// ACSIIStyles is a map defining ANSI color styles for different ASCII markers used in formatting output.
var ACSIIStyles = map[string]string{
	ASCIIAdded:   "30;42",
	ASCIIDeleted: "30;41",
}

func (f *ASCIIFormatter) push(name string, size int, array bool) {
	f.path = append(f.path, name)
	f.size = append(f.size, size)
	f.inArray = append(f.inArray, array)
}

func (f *ASCIIFormatter) pop() {
	f.path = f.path[0 : len(f.path)-1]
	f.size = f.size[0 : len(f.size)-1]
	f.inArray = f.inArray[0 : len(f.inArray)-1]
}

func (f *ASCIIFormatter) addLineWith(value string) {
	f.line = &ASCIILine{
		marker: ASCIISame,
		indent: len(f.path),
		buffer: bytes.NewBufferString(value),
	}
	f.closeLine()
}

func (f *ASCIIFormatter) newLine(marker string) {
	f.line = &ASCIILine{
		marker: marker,
		indent: len(f.path),
		buffer: bytes.NewBuffer([]byte{}),
	}
}

func (f *ASCIIFormatter) closeLine() {
	style, ok := ACSIIStyles[f.line.marker]
	if f.config.Coloring && ok {
		f.buffer.WriteString("\x1b[" + style + "m")
	}

	f.buffer.WriteString(f.line.marker)

	for n := 0; n < f.line.indent; n++ {
		f.buffer.WriteString("  ")
	}

	f.buffer.Write(f.line.buffer.Bytes())

	if f.config.Coloring && ok {
		f.buffer.WriteString("\x1b[0m")
	}

	f.buffer.WriteRune('\n')
}

func (f *ASCIIFormatter) printKey(name string) {
	if !f.inArray[len(f.inArray)-1] {
		fmt.Fprintf(f.line.buffer, `"%s": `, name)
	} else if f.config.ShowArrayIndex {
		fmt.Fprintf(f.line.buffer, `%s: `, name)
	}
}

func (f *ASCIIFormatter) printComma() {
	f.size[len(f.size)-1]--
	if f.size[len(f.size)-1] > 0 {
		f.line.buffer.WriteRune(',')
	}
}

func (f *ASCIIFormatter) printValue(value interface{}) {
	switch v := value.(type) {
	case uint64:
		fmt.Fprint(f.line.buffer, v)
	case json.Number:
		fmt.Fprint(f.line.buffer, v.String())
	case string:
		fmt.Fprintf(f.line.buffer, `"%s"`, value)
	case nil:
		f.line.buffer.WriteString("null")
	default:
		fmt.Fprintf(f.line.buffer, `%#v`, value)
	}
}

func (f *ASCIIFormatter) print(a string) {
	f.line.buffer.WriteString(a)
}

func (f *ASCIIFormatter) printRecursive(name string, value interface{}, marker string) {
	switch value := value.(type) {
	case map[string]interface{}:
		f.newLine(marker)
		f.printKey(name)
		f.print("{")
		f.closeLine()

		m := value
		size := len(m)
		f.push(name, size, false)

		keys := sortedKeys(m)
		for _, key := range keys {
			f.printRecursive(key, m[key], marker)
		}

		f.pop()

		f.newLine(marker)
		f.print("}")
		f.printComma()
		f.closeLine()

	case []interface{}:
		f.newLine(marker)
		f.printKey(name)
		f.print("[")
		f.closeLine()

		s := value
		size := len(s)
		f.push("", size, true)

		for _, item := range s {
			f.printRecursive("", item, marker)
		}

		f.pop()

		f.newLine(marker)
		f.print("]")
		f.printComma()
		f.closeLine()

	default:
		f.newLine(marker)
		f.printKey(name)
		f.printValue(value)
		f.printComma()
		f.closeLine()
	}
}
