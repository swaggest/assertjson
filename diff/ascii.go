package diff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func NewAsciiFormatter(left interface{}, config AsciiFormatterConfig) *AsciiFormatter {
	return &AsciiFormatter{
		left:   left,
		config: config,
	}
}

type AsciiFormatter struct {
	left    interface{}
	config  AsciiFormatterConfig
	buffer  *bytes.Buffer
	path    []string
	size    []int
	inArray []bool
	line    *AsciiLine
}

type AsciiFormatterConfig struct {
	ShowArrayIndex bool
	Coloring       bool
}

var AsciiFormatterDefaultConfig = AsciiFormatterConfig{}

type AsciiLine struct {
	marker string
	indent int
	buffer *bytes.Buffer
}

func (f *AsciiFormatter) Format(diff Diff) (result string, err error) {
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

func (f *AsciiFormatter) formatObject(left map[string]interface{}, df Diff) {
	f.addLineWith("{")
	f.push("ROOT", len(left), false)
	f.processObject(left, df.Deltas())
	f.pop()
	f.addLineWith("}")
}

func (f *AsciiFormatter) formatArray(left []interface{}, df Diff) {
	f.addLineWith("[")
	f.push("ROOT", len(left), true)
	f.processArray(left, df.Deltas())
	f.pop()
	f.addLineWith("]")
}

func (f *AsciiFormatter) processArray(array []interface{}, deltas []Delta) error {
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

			f.printRecursive(d.String(), d.Value, AsciiAdded)
		}
	}

	return nil
}

func (f *AsciiFormatter) processObject(object map[string]interface{}, deltas []Delta) error {
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
			f.printRecursive(d.String(), d.Value, AsciiAdded)
		}
	}

	return nil
}

func (f *AsciiFormatter) processItem(value interface{}, deltas []Delta, position Position) error {
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

				f.newLine(AsciiSame)
				f.printKey(positionStr)
				f.print("{")
				f.closeLine()
				f.push(positionStr, len(o), false)
				f.processObject(o, d.Deltas)
				f.pop()
				f.newLine(AsciiSame)
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

				f.newLine(AsciiSame)
				f.printKey(positionStr)
				f.print("[")
				f.closeLine()
				f.push(positionStr, len(a), true)
				f.processArray(a, d.Deltas)
				f.pop()
				f.newLine(AsciiSame)
				f.print("]")
				f.printComma()
				f.closeLine()

			case *Added:
				d := matchedDelta.(*Added)
				f.printRecursive(positionStr, d.Value, AsciiAdded)

				f.size[len(f.size)-1]++

			case *Modified:
				d := matchedDelta.(*Modified)
				savedSize := f.size[len(f.size)-1]
				f.printRecursive(positionStr, d.OldValue, AsciiDeleted)
				f.size[len(f.size)-1] = savedSize
				f.printRecursive(positionStr, d.NewValue, AsciiAdded)

			case *TextDiff:
				savedSize := f.size[len(f.size)-1]
				d := matchedDelta.(*TextDiff)
				f.printRecursive(positionStr, d.OldValue, AsciiDeleted)
				f.size[len(f.size)-1] = savedSize
				f.printRecursive(positionStr, d.NewValue, AsciiAdded)

			case *Deleted:
				d := matchedDelta.(*Deleted)
				f.printRecursive(positionStr, d.Value, AsciiDeleted)

			default:
				return errors.New("Unknown Delta type detected")
			}
		}
	} else {
		f.printRecursive(positionStr, value, AsciiSame)
	}

	return nil
}

func (f *AsciiFormatter) searchDeltas(deltas []Delta, position Position) (results []Delta) {
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
	AsciiSame    = " "
	AsciiAdded   = "+"
	AsciiDeleted = "-"
)

// ACSIIStyles is a map defining ANSI color styles for different ASCII markers used in formatting output.
var ACSIIStyles = map[string]string{
	AsciiAdded:   "30;42",
	AsciiDeleted: "30;41",
}

func (f *AsciiFormatter) push(name string, size int, array bool) {
	f.path = append(f.path, name)
	f.size = append(f.size, size)
	f.inArray = append(f.inArray, array)
}

func (f *AsciiFormatter) pop() {
	f.path = f.path[0 : len(f.path)-1]
	f.size = f.size[0 : len(f.size)-1]
	f.inArray = f.inArray[0 : len(f.inArray)-1]
}

func (f *AsciiFormatter) addLineWith(value string) {
	f.line = &AsciiLine{
		marker: AsciiSame,
		indent: len(f.path),
		buffer: bytes.NewBufferString(value),
	}
	f.closeLine()
}

func (f *AsciiFormatter) newLine(marker string) {
	f.line = &AsciiLine{
		marker: marker,
		indent: len(f.path),
		buffer: bytes.NewBuffer([]byte{}),
	}
}

func (f *AsciiFormatter) closeLine() {
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

func (f *AsciiFormatter) printKey(name string) {
	if !f.inArray[len(f.inArray)-1] {
		fmt.Fprintf(f.line.buffer, `"%s": `, name)
	} else if f.config.ShowArrayIndex {
		fmt.Fprintf(f.line.buffer, `%s: `, name)
	}
}

func (f *AsciiFormatter) printComma() {
	f.size[len(f.size)-1]--
	if f.size[len(f.size)-1] > 0 {
		f.line.buffer.WriteRune(',')
	}
}

func (f *AsciiFormatter) printValue(value interface{}) {
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

func (f *AsciiFormatter) print(a string) {
	f.line.buffer.WriteString(a)
}

func (f *AsciiFormatter) printRecursive(name string, value interface{}, marker string) {
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
