package util

import (
	"bytes"
	"math"
	"sort"
	"strings"

	"godejs/internal/js"
)

func deepStrictEqual(a, b js.Value, skipPrototype bool) bool {
	return equalValues(a, b, skipPrototype, map[[2]js.Value]bool{})
}

func equalValues(a, b js.Value, skipPrototype bool, seen map[[2]js.Value]bool) bool {
	if !a.Valid() || !b.Valid() {
		return a.Valid() == b.Valid()
	}
	if a.IsUndefined() || b.IsUndefined() {
		return a.IsUndefined() && b.IsUndefined()
	}
	if a.IsNull() || b.IsNull() {
		return a.IsNull() && b.IsNull()
	}
	if a.IsBool() || b.IsBool() {
		return a.IsBool() && b.IsBool() && a.Bool() == b.Bool()
	}
	if a.IsNumber() || b.IsNumber() {
		if !a.IsNumber() || !b.IsNumber() {
			return false
		}
		x, y := a.Float(), b.Float()
		if math.IsNaN(x) || math.IsNaN(y) {
			return math.IsNaN(x) && math.IsNaN(y)
		}
		if x == 0 && y == 0 {
			return math.Signbit(x) == math.Signbit(y)
		}
		return x == y
	}
	if a.IsString() || b.IsString() {
		return a.IsString() && b.IsString() && a.Str() == b.Str()
	}
	if a.IsFunction() || b.IsFunction() {
		return a == b
	}
	if !a.IsObject() || !b.IsObject() {
		return a == b
	}

	ta, tb := a.Tag(), b.Tag()
	if ta != tb {
		return false
	}
	key := [2]js.Value{a, b}
	if seen[key] {
		return true
	}
	seen[key] = true

	// Node's strict equality requires identical prototypes for every object
	// unless the new skipPrototype argument explicitly opts out. This is what
	// makes a Uint8Array differ from a Buffer and a class instance differ from
	// a plain object with the same fields.
	if !skipPrototype {
		pa, pb := a.Object().Proto(), b.Object().Proto()
		if pa.Valid() != pb.Valid() {
			return false
		}
		if pa.Valid() && pa.V() != pb.V() {
			return false
		}
	}

	switch ta {
	case "Date":
		x, y := dateTime(a), dateTime(b)
		if math.IsNaN(x) || math.IsNaN(y) {
			return math.IsNaN(x) && math.IsNaN(y)
		}
		return x == y
	case "RegExp":
		return jsToString(a) == jsToString(b)
	case "Promise":
		return a == b
	case "Map":
		return equalMaps(a, b, skipPrototype, seen)
	case "Set":
		return equalSets(a, b, skipPrototype, seen)
	case "ArrayBuffer", "SharedArrayBuffer":
		return a.Object().Get("byteLength").Int() == b.Object().Get("byteLength").Int()
	}
	if strings.HasSuffix(ta, "Error") {
		ao, bo := a.Object(), b.Object()
		if ao.Get("name").Str() != bo.Get("name").Str() ||
			ao.Get("message").Str() != bo.Get("message").Str() {
			return false
		}
		return equalOwnProperties(a, b, skipPrototype, seen)
	}
	if typedArrayTags[ta] {
		if !equalTypedArrays(a, b, seen, skipPrototype) {
			return false
		}
		return equalSymbolProperties(a, b, skipPrototype, seen)
	}
	if boxedTag(ta) {
		// Boxed primitives compare their internal slot first: `new Number(2)`
		// and `new Number(1)` have no own keys, so a naive property comparison
		// would consider them equal.
		va, oka := boxedValue(a)
		vb, okb := boxedValue(b)
		if !oka || !okb {
			return false
		}
		if !equalValues(va, vb, skipPrototype, seen) {
			return false
		}
		return equalOwnProperties(a, b, skipPrototype, seen)
	}

	if a.IsArray() || b.IsArray() {
		if !a.IsArray() || !b.IsArray() {
			return false
		}
		oa, ob := a.Object(), b.Object()
		if oa.Length() != ob.Length() {
			return false
		}
		for i := 0; i < oa.Length(); i++ {
			if !equalValues(oa.At(i), ob.At(i), skipPrototype, seen) {
				return false
			}
		}
		return equalSymbolProperties(a, b, skipPrototype, seen)
	}

	return equalOwnProperties(a, b, skipPrototype, seen)
}

func boxedTag(tag string) bool {
	switch tag {
	case "Number", "String", "Boolean", "Symbol", "BigInt":
		return true
	}
	return false
}

func boxedValue(v js.Value) (js.Value, bool) {
	o := v.Object()
	valueOf := o.Get("valueOf")
	if !valueOf.IsFunction() {
		return js.Value{}, false
	}
	r, err := valueOf.Object().Call(o)
	if err != nil {
		return js.Value{}, false
	}
	return r, true
}

func equalOwnProperties(a, b js.Value, skipPrototype bool, seen map[[2]js.Value]bool) bool {
	ao, bo := a.Object(), b.Object()
	ka := ao.OwnEnumerableKeys()
	kb := bo.OwnEnumerableKeys()
	if len(ka) != len(kb) {
		return false
	}
	sort.Strings(ka)
	sort.Strings(kb)
	for i := range ka {
		if ka[i] != kb[i] {
			return false
		}
	}
	for _, k := range ka {
		if !equalValues(ao.Get(k), bo.Get(k), skipPrototype, seen) {
			return false
		}
	}
	return equalSymbolProperties(a, b, skipPrototype, seen)
}

// equalSymbolProperties matches enumerable symbol properties by symbol
// identity. Two symbols with the same description are not interchangeable, so
// the comparison cannot be keyed on their string form.
func equalSymbolProperties(a, b js.Value, skipPrototype bool, seen map[[2]js.Value]bool) bool {
	sa := symbolProps(a.Object())
	sb := symbolProps(b.Object())
	if len(sa) != len(sb) {
		return false
	}
	used := make([]bool, len(sb))
	for _, e := range sa {
		found := false
		for i, o := range sb {
			if used[i] {
				continue
			}
			if e.id.Valid() && o.id.Valid() {
				if e.id != o.id {
					continue
				}
			} else if e.key != o.key {
				continue
			}
			if equalValues(e.val, o.val, skipPrototype, seen) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func dateTime(v js.Value) float64 {
	o := v.Object()
	fn := o.Get("getTime")
	if !fn.IsFunction() {
		return math.NaN()
	}
	r, err := fn.Object().Call(o)
	if err != nil {
		return math.NaN()
	}
	return r.Float()
}

func equalTypedArrays(a, b js.Value, seen map[[2]js.Value]bool, skipPrototype bool) bool {
	if ba, ok := a.Bytes(); ok {
		bb, ok2 := b.Bytes()
		return ok2 && bytes.Equal(ba, bb)
	}
	oa, ob := a.Object(), b.Object()
	if oa.Length() != ob.Length() {
		return false
	}
	for i := 0; i < oa.Length(); i++ {
		if !equalValues(oa.At(i), ob.At(i), skipPrototype, seen) {
			return false
		}
	}
	return true
}

type mapEntry struct {
	key js.Value
	val js.Value
}

func collectMap(v js.Value) []mapEntry {
	var out []mapEntry
	o := v.Object()
	f := o.Get("forEach")
	if !f.IsFunction() {
		return out
	}
	cb := makeFunc(v.Ctx(), "callback", func(_ js.Object, args []js.Value) (js.Value, error) {
		var val, key js.Value
		if len(args) > 0 {
			val = args[0]
		}
		if len(args) > 1 {
			key = args[1]
		}
		out = append(out, mapEntry{key: key, val: val})
		return v.Ctx().Undefined(), nil
	})
	_, _ = f.Object().Call(o, cb)
	return out
}

func collectSet(v js.Value) []js.Value {
	var out []js.Value
	o := v.Object()
	f := o.Get("forEach")
	if !f.IsFunction() {
		return out
	}
	cb := makeFunc(v.Ctx(), "callback", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			out = append(out, args[0])
		}
		return v.Ctx().Undefined(), nil
	})
	_, _ = f.Object().Call(o, cb)
	return out
}

func equalMaps(a, b js.Value, skipPrototype bool, seen map[[2]js.Value]bool) bool {
	am, bm := collectMap(a), collectMap(b)
	if len(am) != len(bm) {
		return false
	}
	used := make([]bool, len(bm))
	for _, e := range am {
		found := false
		for i, o := range bm {
			if used[i] {
				continue
			}
			if equalValues(e.key, o.key, skipPrototype, seen) && equalValues(e.val, o.val, skipPrototype, seen) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalSets(a, b js.Value, skipPrototype bool, seen map[[2]js.Value]bool) bool {
	as, bs := collectSet(a), collectSet(b)
	if len(as) != len(bs) {
		return false
	}
	used := make([]bool, len(bs))
	for _, e := range as {
		found := false
		for i, o := range bs {
			if used[i] {
				continue
			}
			if equalValues(e, o, skipPrototype, seen) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
