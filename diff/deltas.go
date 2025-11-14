package diff

import (
	"reflect"
	"strconv"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// A Delta represents an atomic difference between two JSON objects.
type Delta interface {
	// Similarity calculates the similarity of the Delta values.
	// The return value is normalized from 0 to 1,
	// 0 is completely different and 1 is they are same
	Similarity() (similarity float64)
}

// To cache the calculated similarity,
// concrete Deltas can use similariter and similarityCache.
type similariter interface {
	similarity() (similarity float64)
}

type similarityCache struct {
	similariter

	value float64
}

func newSimilarityCache(sim similariter) similarityCache {
	cache := similarityCache{similariter: sim, value: -1}

	return cache
}

func (cache similarityCache) Similarity() (similarity float64) {
	if cache.value < 0 {
		cache.value = cache.similarity()
	}

	return cache.value
}

// A Position represents the position of a Delta in an object or an array.
type Position interface {
	// String returns the position as a string
	String() (name string)

	// CompareTo returns a true if the Position is smaller than another Position.
	// This function is used to sort Positions by the sort package.
	CompareTo(another Position) bool
}

// A Name is a Postition with a string, which means the delta is in an object.
type Name string

// String returns the string representation of the Name.
func (n Name) String() (name string) {
	return string(n)
}

// CompareTo returns true if the Name is lexicographically less than the given Position, which must be of type Name.
func (n Name) CompareTo(another Position) bool {
	return n < another.(Name)
}

// Index is a Position with an int value, which means the Delta is in an Array.
type Index int

// String converts the Index value to its string representation and returns it.
func (i Index) String() (name string) {
	return strconv.Itoa(int(i))
}

// CompareTo compares the current Index with another Position and returns true if the current Index is smaller.
func (i Index) CompareTo(another Position) bool {
	return i < another.(Index)
}

// A PreDelta is a Delta that has a position of the left side JSON object.
// Deltas implements this interface should be applies before PostDeltas.
type PreDelta interface {
	// PrePosition returns the Position.
	PrePosition() Position

	// PreApply applies the delta to object.
	PreApply(object interface{}) interface{}
}

type preDelta struct{ Position }

func (i preDelta) PrePosition() Position {
	return i.Position
}

type preDeltas []PreDelta

// Len returns the number of elements in the preDeltas collection.
func (s preDeltas) Len() int {
	return len(s)
}

// Swap exchanges the elements at indices i and j in the preDeltas collection.
func (s preDeltas) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the element with index i should sort before the element with index j in the preDeltas collection.
func (s preDeltas) Less(i, j int) bool {
	return !s[i].PrePosition().CompareTo(s[j].PrePosition())
}

// PostDelta represents an interface for deltas applied post object modification.
// It requires methods to retrieve the position and apply the delta to an object.
type PostDelta interface {
	// PostPosition returns the Position.
	PostPosition() Position

	// PostApply applies the delta to object.
	PostApply(object interface{}) interface{}
}

type postDelta struct{ Position }

func (i postDelta) PostPosition() Position {
	return i.Position
}

type postDeltas []PostDelta

// Len returns the number of elements in the postDeltas slice. It is used for implementing the sort.Interface.
func (s postDeltas) Len() int {
	return len(s)
}

// Swap swaps the elements at indices i and j in the postDeltas slice. It is used to implement the sort.Interface.
func (s postDeltas) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less compares the PostPosition of elements at indices i and j and returns true if the i-th element is less than the j-th.
func (s postDeltas) Less(i, j int) bool {
	return s[i].PostPosition().CompareTo(s[j].PostPosition())
}

// An Object is a Delta that represents an object of JSON.
type Object struct {
	postDelta
	similarityCache

	// Deltas holds internal Deltas
	Deltas []Delta
}

// NewObject returns an Object.
func NewObject(position Position, deltas []Delta) *Object {
	d := Object{postDelta: postDelta{position}, Deltas: deltas}
	d.similarityCache = newSimilarityCache(&d)

	return &d
}

// PostApply processes the given object by applying deltas at positions determined by the object's type (map or slice).
func (d *Object) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		n := string(d.PostPosition().(Name))
		o[n] = applyDeltas(d.Deltas, o[n])
	case []interface{}:
		n := int(d.PostPosition().(Index))
		o[n] = applyDeltas(d.Deltas, o[n])
	}

	return object
}

func (d *Object) similarity() (similarity float64) {
	similarity = deltasSimilarity(d.Deltas)

	return similarity
}

// An Array is a Delta that represents an array of JSON.
type Array struct {
	postDelta
	similarityCache

	// Deltas holds internal Deltas
	Deltas []Delta
}

// NewArray returns an Array.
func NewArray(position Position, deltas []Delta) *Array {
	d := Array{postDelta: postDelta{position}, Deltas: deltas}
	d.similarityCache = newSimilarityCache(&d)

	return &d
}

// PostApply applies the stored deltas to the provided object based on their positions, modifying and returning the object.
func (d *Array) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		n := string(d.PostPosition().(Name))
		o[n] = applyDeltas(d.Deltas, o[n])
	case []interface{}:
		n := int(d.PostPosition().(Index))
		o[n] = applyDeltas(d.Deltas, o[n])
	}

	return object
}

func (d *Array) similarity() (similarity float64) {
	similarity = deltasSimilarity(d.Deltas)

	return similarity
}

// An Added represents a new added field of an object or an array.
type Added struct {
	postDelta
	similarityCache

	// Values holds the added value
	Value interface{}
}

// NewAdded returns a new Added.
func NewAdded(position Position, value interface{}) *Added {
	d := Added{postDelta: postDelta{position}, Value: value}

	return &d
}

// PostApply applies the added value to the given object at the position specified by the PostPosition method.
func (d *Added) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		object.(map[string]interface{})[string(d.PostPosition().(Name))] = d.Value
	case []interface{}:
		i := int(d.PostPosition().(Index))

		if i < len(o) {
			o = append(o, 0) // dummy
			copy(o[i+1:], o[i:])
			o[i] = d.Value

			return o
		}

		return append(o, d.Value)
	}

	return object
}

func (d *Added) similarity() (similarity float64) {
	return 0
}

// A Modified represents a field whose value is changed.
type Modified struct {
	postDelta
	similarityCache

	// The value before modification
	OldValue interface{}

	// The value after modification
	NewValue interface{}
}

// NewModified returns a Modified.
func NewModified(position Position, oldValue, newValue interface{}) *Modified {
	d := Modified{
		postDelta: postDelta{position},
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	d.similarityCache = newSimilarityCache(&d)

	return &d
}

// PostApply updates a map or slice at a specific position with a new value and returns the modified object.
func (d *Modified) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		o[string(d.PostPosition().(Name))] = d.NewValue
	case []interface{}:
		o[(d.PostPosition().(Index))] = d.NewValue
	}

	return object
}

func (d *Modified) similarity() (similarity float64) {
	similarity += 0.3 // at least, they are at the same position
	if reflect.TypeOf(d.OldValue) == reflect.TypeOf(d.NewValue) {
		similarity += 0.3 // types are same

		switch t := d.OldValue.(type) {
		case string:
			similarity += 0.4 * stringSimilarity(t, d.NewValue.(string))
		case float64:
			ratio := t / d.NewValue.(float64)
			if ratio > 1 {
				ratio = 1 / ratio
			}

			similarity += 0.4 * ratio
		}
	}

	return similarity
}

// A TextDiff represents a Modified with TextDiff between the old and the new values.
type TextDiff struct {
	Modified

	// Diff string
	Diff []dmp.Patch
}

// NewTextDiff creates a new TextDiff instance with the provided position, diff, oldValue, and newValue.
func NewTextDiff(position Position, diff []dmp.Patch, oldValue, newValue interface{}) *TextDiff {
	d := TextDiff{
		Modified: *NewModified(position, oldValue, newValue),
		Diff:     diff,
	}

	return &d
}

// PostApply updates the provided object with the changes specified in the TextDiff and returns the modified object.
func (d *TextDiff) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		i := string(d.PostPosition().(Name))
		d.OldValue = o[i]
		d.patch()
		o[i] = d.NewValue
	case []interface{}:
		i := d.PostPosition().(Index)
		d.OldValue = o[i]
		d.patch()
		o[i] = d.NewValue
	}

	return object
}

func (d *TextDiff) patch() {
	if d.OldValue == nil {
		panic("old Value is not set")
	}

	patcher := dmp.New()

	patched, successes := patcher.PatchApply(d.Diff, d.OldValue.(string))
	for _, success := range successes {
		if !success {
			panic("failed to apply a patch")
		}
	}

	d.NewValue = patched
}

// DiffString returns the textual representation of the diff stored in the TextDiff instance.
func (d *TextDiff) DiffString() string {
	dm := dmp.New()

	return dm.PatchToText(d.Diff)
}

// Deleted represents a change where an element is removed from a map or slice at a specific position.
// It embeds preDelta to store positional metadata and includes the Value field to reference the deleted element.
type Deleted struct {
	preDelta

	// The value deleted
	Value interface{}
}

// NewDeleted returns a Deleted.
func NewDeleted(position Position, value interface{}) *Deleted {
	d := Deleted{
		preDelta: preDelta{position},
		Value:    value,
	}

	return &d
}

// PreApply removes an element from a map or slice based on the position specified in the Deleted instance.
func (d Deleted) PreApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		delete(object.(map[string]interface{}), string(d.PrePosition().(Name)))
	case []interface{}:
		i := int(d.PrePosition().(Index))

		return append(o[:i], o[i+1:]...)
	}

	return object
}

// Similarity calculates and returns the similarity for the Deleted delta type as a floating-point value.
func (d Deleted) Similarity() (similarity float64) {
	return 0
}

// A Moved represents field that is moved, which means the index or name is
// changed. Note that, in this library, assigning a Moved and a Modified to
// a single position is not allowed. For the compatibility with jsondiffpatch,
// the Moved in this library can hold the old and new value in it.
type Moved struct {
	preDelta
	postDelta
	similarityCache

	// The value before moving
	Value interface{}
	// The delta applied after moving (for compatibility)
	Delta interface{}
}

// NewMoved creates and returns a new Moved instance representing a field that has been relocated with old and new positions.
func NewMoved(oldPosition Position, newPosition Position, value interface{}, delta Delta) *Moved {
	d := Moved{
		preDelta:  preDelta{oldPosition},
		postDelta: postDelta{newPosition},
		Value:     value,
		Delta:     delta,
	}
	d.similarityCache = newSimilarityCache(&d)

	return &d
}

// PreApply modifies the given object by removing the element at the pre-move index and storing its value in the Moved instance.
func (d *Moved) PreApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		// not supported
	case []interface{}:
		i := int(d.PrePosition().(Index))
		d.Value = o[i]

		return append(o[:i], o[i+1:]...)
	}

	return object
}

// PostApply applies the stored delta after a move operation and adjusts the position of a value in the object.
func (d *Moved) PostApply(object interface{}) interface{} {
	switch o := object.(type) {
	case map[string]interface{}:
		// not supported
	case []interface{}:
		i := int(d.PostPosition().(Index))

		o = append(o, 0) // dummy
		copy(o[i+1:], o[i:])
		o[i] = d.Value
		object = o
	}

	if d.Delta != nil {
		d.Delta.(PostDelta).PostApply(object)
	}

	return object
}

func (d *Moved) similarity() (similarity float64) {
	similarity = 0.6 // as type and contents are same

	ratio := float64(d.PrePosition().(Index)) / float64(d.PostPosition().(Index))
	if ratio > 1 {
		ratio = 1 / ratio
	}

	similarity += 0.4 * ratio

	return similarity
}
