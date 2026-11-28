// Package protonum translates generated protobuf enums to the names a
// service stores, and back.
//
// Each enum gets one Translator, built when its package loads. New panics
// unless every defined value but the zero value has a name of its own, so
// a value added to the proto without a name fails every binary and test
// that imports the translator, not one request in production.
package protonum

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ErrUnnamed is the error Names wraps for a value with no name: the zero
// value, or a number the enum does not define. For a request, it is the
// caller's mistake.
var ErrUnnamed = errors.New("protonum: value has no name")

// ErrUnknownName is the error Enums wraps for a name no value has. For
// stored data, the data and the enum have drifted apart.
var ErrUnknownName = errors.New("protonum: no value has the name")

// Enum is a generated protobuf enum.
type Enum interface {
	~int32
	protoreflect.Enum
}

// Translator maps one protobuf enum E to its stored names N and back.
// N is any string type, so a domain package can keep its own name type.
// It maps names, never numbers: enum numbers are a wire contract, and a
// table's ids are data.
type Translator[E Enum, N ~string] struct {
	names map[E]N
	enums map[N]E
}

// New returns a Translator over names. It panics when a number is not a
// value of E, when the zero value has a name (the zero value means unset),
// when a name is empty or shared by two values, or when a defined value
// other than zero has no name. New checks values in a fixed order, so the
// same mistake always panics with the same message.
func New[E Enum, N ~string](names map[E]N) Translator[E, N] {
	var zero E
	values := zero.Descriptor().Values()
	enums := make(map[N]E, len(names))

	for _, e := range slices.Sorted(maps.Keys(names)) {
		name := names[e]
		value := values.ByNumber(e.Number())
		switch {
		case value == nil:
			panic(fmt.Sprintf("protonum: %s has no value %d", zero.Descriptor().FullName(), e))
		case e == zero:
			panic(fmt.Sprintf("protonum: %s gets a name", value.FullName()))
		case name == "":
			panic(fmt.Sprintf("protonum: %s has an empty name", value.FullName()))
		}

		if other, taken := enums[name]; taken {
			panic(fmt.Sprintf("protonum: %s and %s share the name %q", values.ByNumber(other.Number()).FullName(), value.FullName(), name))
		}

		enums[name] = e
	}

	for i := range values.Len() {
		v := values.Get(i)
		if _, named := names[E(v.Number())]; v.Number() != 0 && !named {
			panic(fmt.Sprintf("protonum: %s has no name", v.FullName()))
		}
	}

	return Translator[E, N]{names: maps.Clone(names), enums: enums}
}

// Name returns e's name. It reports false for the zero value and for a
// number E does not define.
func (t Translator[E, N]) Name(e E) (N, bool) {
	name, ok := t.names[e]

	return name, ok
}

// NameOf returns e's name, or the empty name for the zero value, which
// means unset, and for an undefined number.
func (t Translator[E, N]) NameOf(e E) N {
	return t.names[e]
}

// Enum returns the value called name. It reports false for a name no value
// has.
func (t Translator[E, N]) Enum(name N) (E, bool) {
	e, ok := t.enums[name]

	return e, ok
}

// Names translates values, such as a request's filter. A value with no
// name fails the whole translation with ErrUnnamed rather than drop out of
// a filter and widen it. No values translate to nil.
func (t Translator[E, N]) Names(es []E) ([]N, error) {
	if len(es) == 0 {
		return nil, nil
	}
	names := make([]N, len(es))
	for i, e := range es {
		name, ok := t.names[e]
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnnamed, e)
		}
		names[i] = name
	}

	return names, nil
}

// KnownNames translates values and leaves out a value with no name, such
// as one a newer peer sends.
func (t Translator[E, N]) KnownNames(es []E) []N {
	var names []N
	for _, e := range es {
		if name, ok := t.names[e]; ok {
			names = append(names, name)
		}
	}

	return names
}

// Enums translates names, such as stored ones. A name no value has fails
// the whole translation with ErrUnknownName. No names translate to nil.
func (t Translator[E, N]) Enums(names []N) ([]E, error) {
	if len(names) == 0 {
		return nil, nil
	}
	es := make([]E, len(names))
	for i, name := range names {
		e, ok := t.enums[name]
		if !ok {
			var zero E

			return nil, fmt.Errorf("%w: %s has no value named %q", ErrUnknownName, zero.Descriptor().FullName(), name)
		}
		es[i] = e
	}

	return es, nil
}

// All returns every name in the order of their enum numbers, such as for a
// test that compares the names with another source of truth.
func (t Translator[E, N]) All() []N {
	es := slices.Sorted(maps.Keys(t.names))
	names := make([]N, len(es))
	for i, e := range es {
		names[i] = t.names[e]
	}

	return names
}
