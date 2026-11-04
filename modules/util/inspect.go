package util

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"godejs/internal/js"
)

type inspectOptions struct {
	depth         int
	depthInfinite bool
	showHidden    bool
	sorted        bool
	customInspect bool
	getters       bool
	// maxArrayLength and maxStringLength cap how much of a long value is
	// rendered; zero means Infinity.
	maxArrayLength  int
	maxStringLength int
	// breakLength is the column at which a container switches to one entry
	// per line. Zero means Infinity (never break).
	breakLength int
	// numericSeparator groups long numbers with underscores, both in inspect
	// output and in util.formatWithOptions placeholders.
	numericSeparator bool
	// compact allows a container whose inline form fits breakLength to stay on
	// one line. compact: false forces one entry per line, which assert's diff
	// renderer relies on.
	compact bool
}

// Node's compact renderer wraps objects slightly before its nominal 80-column
// breakLength; 72 reproduces the observed default breaks (for example a
// six-property object whose inline form is 73 columns).
var defaultInspectOptions = inspectOptions{
	depth:           2,
	breakLength:     72,
	compact:         true,
	customInspect:   true,
	maxArrayLength:  100,
	maxStringLength: 10000,
}

// customRender is the JavaScript half of the `util.inspect.custom` protocol.
// It reports whether a value defines the symbol and, when it does, the string
// the hook produced. customInspectFn is the public util.inspect handed to the
// hook as its third argument.
var (
	customRender    js.Value
	customInspectFn js.Value
)

func parseInspectOptions(v js.Value) inspectOptions {
	o := defaultInspectOptions
	if !v.Valid() || !v.IsObject() {
		return o
	}
	obj := v.Object()
	switch d := obj.Get("depth"); {
	case d.IsNull():
		o.depthInfinite = true
		o.depth = 0
	case d.IsNumber():
		if math.IsInf(d.Float(), 0) {
			o.depthInfinite = true
			o.depth = 0
			break
		}
		// A negative depth is meaningful: it collapses every non-empty
		// container, so it must reach the renderer, not clamp to zero.
		o.depth = d.Int()
	}
	if b := obj.Get("breakLength"); b.IsNumber() {
		f := b.Float()
		if math.IsInf(f, 0) || f <= 0 {
			o.breakLength = 0
		} else {
			o.breakLength = b.Int()
		}
	}
	if n := obj.Get("maxArrayLength"); n.IsNumber() {
		if math.IsInf(n.Float(), 0) {
			o.maxArrayLength = 0
		} else if n.Int() > 0 {
			o.maxArrayLength = n.Int()
		}
	} else if n.IsNull() {
		o.maxArrayLength = 0
	}
	if n := obj.Get("maxStringLength"); n.IsNumber() {
		if math.IsInf(n.Float(), 0) {
			o.maxStringLength = 0
		} else if n.Int() > 0 {
			o.maxStringLength = n.Int()
		}
	} else if n.IsNull() {
		o.maxStringLength = 0
	}
	o.showHidden = obj.Get("showHidden").Bool()
	o.sorted = obj.Get("sorted").Bool()
	if c := obj.Get("customInspect"); c.IsBool() {
		o.customInspect = c.Bool()
	}
	o.getters = obj.Get("getters").Bool()
	if sep := obj.Get("numericSeparator"); sep.IsBool() {
		o.numericSeparator = sep.Bool()
	}
	if c := obj.Get("compact"); c.IsBool() {
		o.compact = c.Bool()
	} else if c.IsNumber() {
		o.compact = c.Int() > 0
	}
	return o
}

type inspectState struct {
	opts    inspectOptions
	indices map[js.Value]int
	next    int
	stack   []js.Value
	cycles  map[js.Value]int
}

func inspectWith(v js.Value, o inspectOptions) string {
	st := &inspectState{
		opts:    o,
		indices: map[js.Value]int{},
		cycles:  map[js.Value]int{},
	}
	// render handles an exhausted depth itself (collapsing only non-empty
	// containers), so even depth zero starts at the top level, like Node.
	return st.render(v, o.depth, 0)
}

func markerFor(v js.Value) string {
	if v.IsArray() {
		return "[Array]"
	}
	switch v.Tag() {
	case "Map":
		return "[Map]"
	case "Set":
		return "[Set]"
	}
	o := v.Object()
	if o.IsNullPrototype() {
		return "[Object: null prototype]"
	}
	// Class instances keep their name (`[Foo]`), arguments collapse to a
	// bare object marker, exactly like Node.
	if name := constructorName(o); name != "" && name != "Object" {
		return "[" + name + "]"
	}
	return "[Object]"
}

// isEmptyContainer reports whether there is nothing to hide: Node only
// collapses a value whose depth is exhausted when it has content, so empty
// containers fall through to the normal render (which spells every empty
// form, including class and null-prototype prefixes, already).
func (st *inspectState) isEmptyContainer(v js.Value) bool {
	o := v.Object()
	switch {
	case v.IsArray():
		return o.Length() == 0 && len(o.OwnEnumerableKeys()) == 0
	case v.Tag() == "Map" || v.Tag() == "Set":
		return o.Get("size").Int() == 0
	default:
		return len(o.OwnEnumerableKeys()) == 0 && len(st.symbolEntries(o)) == 0
	}
}

// Inspect is util.inspect with default options.
func Inspect(v js.Value) string { return inspectWith(v, defaultInspectOptions) }

var typedArrayTags = map[string]bool{
	"Int8Array":         true,
	"Int16Array":        true,
	"Int32Array":        true,
	"Uint8Array":        true,
	"Uint8ClampedArray": true,
	"Uint16Array":       true,
	"Uint32Array":       true,
	"Float16Array":      true,
	"Float32Array":      true,
	"Float64Array":      true,
	"BigInt64Array":     true,
	"BigUint64Array":    true,
}

// render returns the representation of v starting at column indent. indent is
// used only when the value breaks across lines.
func (st *inspectState) render(v js.Value, depth, indent int) string {
	if !v.Valid() {
		return "undefined"
	}
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		if v.Bool() {
			return "true"
		}
		return "false"
	case v.IsNumber():
		return formatNumberWith(v, st.opts.numericSeparator)
	case v.IsString():
		return quoteString(st.truncateString(v.Str()))
	case v.IsFunction():
		return formatFunction(v)
	}
	if !v.IsObject() {
		switch v.Tag() {
		case "Symbol":
			return jsToString(v)
		case "BigInt":
			return jsToString(v) + "n"
		}
		if s, ok := v.StrOK(); ok {
			return quoteString(s)
		}
		return "undefined"
	}

	o := v.Object()
	// Buffer.prototype[util.inspect.custom] renders the Node form including
	// any extra own properties, so it must win over the plain byte dump.
	if st.opts.customInspect {
		if text, ok := st.custom(v, depth, indent); ok {
			return text
		}
	}
	if b, ok := v.Bytes(); ok && o.Get("_godeBuffer").Bool() {
		return formatBuffer(v, b)
	}
	tag := v.Tag()
	switch tag {
	case "Date":
		return formatDate(o)
	case "RegExp":
		return jsToString(v)
	case "ArrayBuffer", "SharedArrayBuffer":
		return tag + " { byteLength: " + formatElement(o.Get("byteLength")) + " }"
	case "Promise":
		return "Promise { <pending> }"
	case "WeakMap":
		return "WeakMap { <items unknown> }"
	case "WeakSet":
		return "WeakSet { <items unknown> }"
	case "Number", "String", "Boolean", "Symbol", "BigInt":
		return st.formatBoxed(o, tag, depth, indent)
	}
	if strings.HasSuffix(tag, "Error") {
		return formatError(o)
	}
	if typedArrayTags[tag] {
		return formatTypedArray(o, tag)
	}

	// Circularity wins over depth collapsing, as in Node: a back-reference at
	// an exhausted depth still prints [Circular *n], not the class marker.
	// Checking the stack without entering keeps depth-collapsed values from
	// consuming reference numbers.
	if idx, ok := st.stackIndex(v); ok {
		// Only the value actually referenced again gets the <ref *n> prefix;
		// marking every object on the stack would tag innocent ancestors
		// (for example an object reached through a getter).
		st.cycles[v] = idx
		return "[Circular *" + strconv.Itoa(idx) + "]"
	}
	if !st.opts.depthInfinite && depth < 0 && !st.isEmptyContainer(v) {
		return markerFor(v)
	}

	st.enter(v)
	var text string
	switch {
	case v.IsArray():
		text = st.renderArray(o, depth, indent)
	case tag == "Map":
		text = st.renderMap(o, depth, indent)
	case tag == "Set":
		text = st.renderSet(o, depth, indent)
	default:
		text = st.renderObject(o, depth, indent)
	}
	st.leave()
	if n, ok := st.cycles[v]; ok {
		text = "<ref *" + strconv.Itoa(n) + "> " + text
	}
	return text
}

// fits reports whether an inline representation starting at indent fits within
// breakLength.
func (st *inspectState) fits(inline string, indent int) bool {
	if st.opts.breakLength <= 0 {
		return true
	}
	return indent+len(inline) <= st.opts.breakLength
}

// stackIndex reports the reference number of a value already being rendered
// higher up the stack.
func (st *inspectState) stackIndex(v js.Value) (int, bool) {
	for _, s := range st.stack {
		if s == v {
			return st.indices[v], true
		}
	}
	return 0, false
}

func (st *inspectState) enter(v js.Value) (int, bool) {
	for _, s := range st.stack {
		if s == v {
			return st.indices[v], true
		}
	}
	st.next++
	idx := st.next
	st.indices[v] = idx
	st.stack = append(st.stack, v)
	return idx, false
}

func (st *inspectState) leave() {
	if len(st.stack) > 0 {
		st.stack = st.stack[:len(st.stack)-1]
	}
}

func (st *inspectState) markCycle() {
	// Retained for callers that need to tag the whole active stack; the
	// renderer itself marks only the value that was found again.
	for _, s := range st.stack {
		if idx, ok := st.indices[s]; ok {
			st.cycles[s] = idx
		}
	}
}

// bufferInspectMax reports buffer.INSPECT_MAX_BYTES through the global the
// buffer module maintains, so assigning it is visible to util.inspect at once.
func bufferInspectMax(v js.Value) int {
	if v.Valid() {
		if n := v.Ctx().Global().Get("__godeInspectMaxBytes"); n.Valid() && n.IsNumber() {
			if f := int(n.Float()); f >= 0 {
				return f
			}
		}
	}
	return 50
}

func formatBuffer(v js.Value, b []byte) string {
	max := bufferInspectMax(v)
	if len(b) == 0 {
		return "<Buffer >"
	}
	parts := make([]string, 0, max)
	n := len(b)
	if n > max {
		n = max
	}
	for i := 0; i < n; i++ {
		parts = append(parts, fmt.Sprintf("%02x", b[i]))
	}
	if len(b) > max {
		more := len(b) - max
		unit := " bytes"
		if more == 1 {
			unit = " byte"
		}
		return "<Buffer " + strings.Join(parts, " ") + " ... " +
			strconv.Itoa(more) + " more" + unit + ">"
	}
	return "<Buffer " + strings.Join(parts, " ") + ">"
}

func formatTypedArray(o js.Object, tag string) string {
	n := o.Length()
	if n == 0 {
		return tag + "(0) []"
	}
	parts := make([]string, 0, n)
	for i := 0; i < n; i++ {
		elem := o.At(i)
		if strings.HasPrefix(tag, "BigInt") {
			parts = append(parts, jsToString(elem)+"n")
		} else {
			parts = append(parts, formatElement(elem))
		}
	}
	return fmt.Sprintf("%s(%d) [ %s ]", tag, n, strings.Join(parts, ", "))
}

func (st *inspectState) renderArray(o js.Object, depth, indent int) string {
	total := o.Length()
	n := total
	if st.opts.maxArrayLength > 0 && n > st.opts.maxArrayLength {
		n = st.opts.maxArrayLength
	}
	if total == 0 {
		return "[]"
	}
	inlineParts := make([]string, 0, n)
	for i := 0; i < n; {
		if !hasOwn(o, strconv.Itoa(i)) {
			j := i
			for j < n && !hasOwn(o, strconv.Itoa(j)) {
				j++
			}
			count := j - i
			if count == 1 {
				inlineParts = append(inlineParts, "<1 empty item>")
			} else {
				inlineParts = append(inlineParts, fmt.Sprintf("<%d empty items>", count))
			}
			i = j
			continue
		}
		inlineParts = append(inlineParts, st.render(o.At(i), depth-1, indent))
		i++
	}
	if n < total {
		inlineParts = append(inlineParts, fmt.Sprintf("... %d more item%s", total-n, plural(total-n)))
	}
	if st.opts.showHidden && n > 0 {
		// Node's %o renders with showHidden, which surfaces the array's
		// length as a bracketed pseudo-property.
		inlineParts = append(inlineParts, fmt.Sprintf("[length]: %d", total))
	}
	if len(inlineParts) == 0 {
		return "[]"
	}
	inline := "[ " + strings.Join(inlineParts, ", ") + " ]"
	if st.opts.compact && st.fits(inline, indent) {
		return inline
	}
	lines := make([]string, 0, n)
	for i := 0; i < n; {
		if !hasOwn(o, strconv.Itoa(i)) {
			j := i
			for j < n && !hasOwn(o, strconv.Itoa(j)) {
				j++
			}
			placeholder := "<1 empty item>"
			if j-i > 1 {
				placeholder = fmt.Sprintf("<%d empty items>", j-i)
			}
			lines = append(lines, strings.Repeat(" ", indent+2)+placeholder)
			i = j
			continue
		}
		lines = append(lines, strings.Repeat(" ", indent+2)+st.render(o.At(i), depth-1, indent+2))
		i++
	}
	if n < total {
		lines = append(lines, strings.Repeat(" ", indent+2)+fmt.Sprintf("... %d more item%s", total-n, plural(total-n)))
	}
	if st.opts.showHidden && n > 0 {
		lines = append(lines, strings.Repeat(" ", indent+2)+fmt.Sprintf("[length]: %d", total))
	}
	return "[\n" + strings.Join(lines, ",\n") + "\n" + strings.Repeat(" ", indent) + "]"
}

func (st *inspectState) renderObject(o js.Object, depth, indent int) string {
	keys := o.OwnEnumerableKeys()
	if st.opts.showHidden {
		keys = o.OwnKeys()
	}
	if st.opts.sorted {
		sort.Strings(keys)
	}
	type entry struct{ key, val string }
	entries := make([]entry, 0, len(keys))
	for _, k := range keys {
		key := k
		if !validIdentifier(k) {
			key = quoteString(k)
		}
		entries = append(entries, entry{key: key, val: st.propText(o, k, depth-1, indent)})
	}
	// Node lists symbol-keyed properties after the string keys.
	for _, sym := range st.symbolEntries(o) {
		entries = append(entries, entry{key: sym.key, val: st.render(sym.val, depth-1, indent)})
	}
	// showHidden surfaces accessors declared on the prototype chain (class
	// getters), and getters: true evaluates them with the instance as the
	// receiver. Own accessors are already covered by propText above.
	if st.opts.getters && st.opts.showHidden {
		seen := make(map[string]bool, len(keys))
		for _, k := range keys {
			seen[k] = true
		}
		for _, acc := range protoAccessors(o) {
			if seen[acc.key] {
				continue
			}
			seen[acc.key] = true
			entries = append(entries, entry{
				key: "[" + acc.key + "]",
				val: st.accessorText(o, acc.desc, depth, indent),
			})
		}
	}

	body := "{}"
	if len(entries) > 0 {
		inlineParts := make([]string, 0, len(entries))
		for _, e := range entries {
			inlineParts = append(inlineParts, e.key+": "+e.val)
		}
		inline := "{ " + strings.Join(inlineParts, ", ") + " }"
		if st.opts.compact && st.fits(inline, indent) {
			body = inline
		} else {
			lines := make([]string, 0, len(entries))
			for _, e := range entries {
				lines = append(lines, strings.Repeat(" ", indent+2)+e.key+": "+e.val)
			}
			body = "{\n" + strings.Join(lines, ",\n") + "\n" + strings.Repeat(" ", indent) + "}"
		}
	}
	if o.IsNullPrototype() {
		return "[Object: null prototype] " + body
	}
	// Class instances are prefixed with their constructor name (`Foo { a: 1 }`),
	// while plain objects keep the bare braces Node shows.
	if name := constructorName(o); name != "" && name != "Object" {
		return name + " " + body
	}
	if v := o.V(); v.Tag() == "Arguments" {
		return "[Arguments] " + body
	}
	return body
}

// accessorInfo is one accessor found on an object's prototype chain.
type accessorInfo struct {
	key  string
	desc js.Value
}

// protoAccessors collects getters and setters declared on the prototype chain,
// nearest prototype first, without repeating a name shadowed lower down.
func protoAccessors(o js.Object) []accessorInfo {
	gopn := o.ObjectCtor().Get("getOwnPropertyNames")
	if !gopn.IsFunction() {
		return nil
	}
	var out []accessorInfo
	seen := map[string]bool{}
	objProto := o.ObjectCtor().Get("prototype")
	proto := o.Proto()
	for hops := 0; hops < 16 && proto.Valid() && proto.IsObject(); hops++ {
		// Object.prototype carries the __proto__ accessor; Node does not print
		// it for every instance, so the chain stops below it.
		if objProto.Valid() && proto.Same(objProto) {
			break
		}
		r, err := gopn.Object().Call(o.ObjectCtor(), proto.V())
		if err != nil || !r.IsObject() {
			break
		}
		arr := r.Object()
		for i := 0; i < arr.Length(); i++ {
			key := arr.At(i).Str()
			if key == "constructor" || seen[key] {
				continue
			}
			desc := propertyDescriptor(proto, key)
			if !desc.Valid() || !desc.IsObject() {
				continue
			}
			d := desc.Object()
			if d.Get("get").IsFunction() || d.Get("set").IsFunction() {
				seen[key] = true
				out = append(out, accessorInfo{key: key, desc: desc})
			}
		}
		proto = proto.Proto()
	}
	return out
}

// accessorText renders an accessor. Node evaluates the getter when getters is
// enabled and shows its value, using the bare marker for objects (which may be
// circular and are rendered by the normal machinery).
func (st *inspectState) accessorText(recv js.Object, desc js.Value, depth, indent int) string {
	d := desc.Object()
	getter := d.Get("get")
	setter := d.Get("set")
	switch {
	case getter.IsFunction() && setter.IsFunction():
		return "[Getter/Setter]"
	case !getter.IsFunction():
		return "[Setter]"
	case !st.opts.getters:
		return "[Getter]"
	}
	v, err := getter.Object().Call(recv)
	if err != nil {
		return "[Getter]"
	}
	if v.IsObject() || v.IsFunction() {
		return "[Getter] " + st.render(v, depth, indent)
	}
	return "[Getter: " + formatElement(v) + "]"
}

func (st *inspectState) propText(o js.Object, key string, depth, indent int) string {
	desc := propertyDescriptor(o, key)
	getter, setter := false, false
	if desc.Valid() && desc.IsObject() {
		d := desc.Object()
		getter = d.Get("get").IsFunction()
		setter = d.Get("set").IsFunction()
	}
	switch {
	case getter && setter:
		return "[Getter/Setter]"
	case getter:
		return "[Getter]"
	case setter:
		return "[Setter]"
	}
	return st.render(o.Get(key), depth, indent)
}

func (st *inspectState) renderMap(o js.Object, depth, indent int) string {
	size := o.Get("size").Int()
	if size == 0 {
		return "Map(0) {}"
	}
	prefix := fmt.Sprintf("Map(%d) ", size)
	var inlineParts []string
	var values [][2]js.Value
	st.forEach(o, func(val, key js.Value) {
		values = append(values, [2]js.Value{key, val})
		inlineParts = append(inlineParts, st.render(key, depth-1, indent)+" => "+st.render(val, depth-1, indent))
	})
	if len(inlineParts) == 0 {
		return fmt.Sprintf("Map(%d) {}", size)
	}
	inline := prefix + "{ " + strings.Join(inlineParts, ", ") + " }"
	if st.opts.compact && st.fits(inline, indent) {
		return inline
	}
	lines := make([]string, 0, len(values))
	for _, pair := range values {
		lines = append(lines, strings.Repeat(" ", indent+2)+st.render(pair[0], depth-1, indent+2)+
			" => "+st.render(pair[1], depth-1, indent+2))
	}
	return prefix + "{\n" + strings.Join(lines, ",\n") + "\n" + strings.Repeat(" ", indent) + "}"
}

func (st *inspectState) renderSet(o js.Object, depth, indent int) string {
	size := o.Get("size").Int()
	if size == 0 {
		return "Set(0) {}"
	}
	prefix := fmt.Sprintf("Set(%d) ", size)
	var inlineParts []string
	var values []js.Value
	st.forEach(o, func(val, _ js.Value) {
		values = append(values, val)
		inlineParts = append(inlineParts, st.render(val, depth-1, indent))
	})
	if len(inlineParts) == 0 {
		return fmt.Sprintf("Set(%d) {}", size)
	}
	inline := prefix + "{ " + strings.Join(inlineParts, ", ") + " }"
	if st.opts.compact && st.fits(inline, indent) {
		return inline
	}
	lines := make([]string, 0, len(values))
	for _, val := range values {
		lines = append(lines, strings.Repeat(" ", indent+2)+st.render(val, depth-1, indent+2))
	}
	return prefix + "{\n" + strings.Join(lines, ",\n") + "\n" + strings.Repeat(" ", indent) + "}"
}

func (st *inspectState) forEach(o js.Object, fn func(a, b js.Value)) {
	f := o.Get("forEach")
	if !f.IsFunction() {
		return
	}
	cb := makeFunc(o.Ctx(), "callback", func(_ js.Object, args []js.Value) (js.Value, error) {
		var a, b js.Value
		if len(args) > 0 {
			a = args[0]
		}
		if len(args) > 1 {
			b = args[1]
		}
		fn(a, b)
		return o.Ctx().Undefined(), nil
	})
	_, _ = f.Object().Call(o, cb)
}

func (st *inspectState) formatBoxed(o js.Object, tag string, depth, indent int) string {
	valueOf := o.Get("valueOf")
	if !valueOf.IsFunction() {
		return "[" + tag + ": null]"
	}
	inner, err := valueOf.Object().Call(o)
	if err != nil {
		return "[" + tag + ": null]"
	}
	if !st.opts.depthInfinite && depth < 0 {
		return "[" + tag + "]"
	}
	return "[" + tag + ": " + st.render(inner, depth-1, indent+len(tag)+3) + "]"
}

func formatDate(o js.Object) string {
	iso := o.Get("toISOString")
	if !iso.IsFunction() {
		return "Invalid Date"
	}
	r, err := iso.Object().Call(o)
	if err != nil {
		return "Invalid Date"
	}
	return r.Str()
}

func formatError(o js.Object) string {
	name := o.Get("name").Str()
	msg := o.Get("message").Str()
	if name == "" {
		name = "Error"
	}
	if msg == "" {
		return name
	}
	return name + ": " + msg
}

func formatFunction(v js.Value) string {
	o := v.Object()
	name := o.Get("name").Str()
	src := ""
	fts := o.Ctx().Global().Get("Function").Object().Get("prototype").Object().Get("toString")
	if fts.IsFunction() {
		// Function.prototype.toString reads its receiver, so pass the
		// function value as `this` rather than an empty object.
		if r, err := fts.Object().Call(o, v); err == nil {
			src = strings.TrimSpace(r.Str())
		}
	}
	if strings.HasPrefix(src, "class") {
		if name == "" {
			return "[class (anonymous)]"
		}
		return "[class " + name + "]"
	}
	// Node names the function flavour, e.g. `[AsyncFunction: foo]`.
	kind := "Function"
	switch {
	case strings.HasPrefix(src, "async function*"), strings.HasPrefix(src, "async function *"):
		kind = "AsyncGeneratorFunction"
	case strings.HasPrefix(src, "function*"):
		kind = "GeneratorFunction"
	case strings.HasPrefix(src, "async "), strings.HasPrefix(src, "async("):
		kind = "AsyncFunction"
	}
	if name == "" {
		return "[" + kind + " (anonymous)]"
	}
	return "[" + kind + ": " + name + "]"
}

func formatNumber(v js.Value) string {
	f := v.Float()
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0 && math.Signbit(f):
		return "-0"
	}
	return jsToString(v)
}

func formatElement(v js.Value) string {
	if !v.Valid() {
		return "undefined"
	}
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		if v.Bool() {
			return "true"
		}
		return "false"
	case v.IsNumber():
		return formatNumber(v)
	case v.IsString():
		return quoteString(v.Str())
	}
	return jsToString(v)
}

func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\v':
			b.WriteString(`\v`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, "\\x%02x", r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func validIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && r != '$' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && r != '$' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func propertyDescriptor(o js.Object, key string) js.Value {
	fn := o.ObjectCtor().Get("getOwnPropertyDescriptor")
	if !fn.IsFunction() {
		return js.Value{}
	}
	r, err := fn.Object().Call(js.Object{}, o.Value, o.Ctx().Str(key))
	if err != nil {
		return js.Value{}
	}
	return r
}

func hasOwn(o js.Object, key string) bool {
	fn := o.ObjectCtor().Get("prototype").Object().Get("hasOwnProperty")
	if !fn.IsFunction() {
		return true
	}
	r, err := fn.Object().Call(o, o.Ctx().Str(key))
	if err != nil {
		return true
	}
	return r.Bool()
}

type symbolEntry struct {
	key string
	val js.Value
	id  js.Value
}

// symbolEntries lists the enumerable symbol-keyed properties, formatted the way
// Node prints them (`[Symbol(desc)]`). jsSymbolEntries is installed by Install.
func (st *inspectState) symbolEntries(o js.Object) []symbolEntry {
	return symbolProps(o)
}

// symbolProps lists the enumerable symbol-keyed properties of o.
func symbolProps(o js.Object) []symbolEntry {
	if !jsSymbolEntries.Valid() {
		return nil
	}
	r, err := jsSymbolEntries.Object().Call(js.Object{}, o.Value)
	if err != nil || !r.IsObject() {
		return nil
	}
	arr := r.Object()
	n := arr.Length()
	if n == 0 {
		return nil
	}
	out := make([]symbolEntry, 0, n)
	for i := 0; i < n; i++ {
		pair := arr.At(i)
		if !pair.IsObject() {
			continue
		}
		p := pair.Object()
		// symbolEntries reports String(symbol) ("Symbol(k)"); Node prints the
		// bracketed form "[Symbol(k)]".
		out = append(out, symbolEntry{key: "[" + p.At(0).Str() + "]", val: p.At(1), id: p.At(2)})
	}
	return out
}

// constructorName returns the name of the object's constructor, or "" when it
// has none (null-prototype objects) or the constructor is anonymous.
func constructorName(o js.Object) string {
	c := o.Get("constructor")
	if !c.Valid() || c.IsUndefined() || c.IsNull() {
		return ""
	}
	if !c.IsObject() {
		return ""
	}
	if n := c.Object().Get("name"); n.IsString() {
		return n.Str()
	}
	return ""
}

// custom runs a value's [util.inspect.custom] hook. It reports false when the
// value has no hook or the hook returned a non-string.
func (st *inspectState) custom(v js.Value, depth, indent int) (string, bool) {
	if !customRender.Valid() || !v.IsObject() {
		return "", false
	}
	c := v.Ctx()
	var depthArg js.Value
	if st.opts.depthInfinite {
		depthArg = c.Null()
	} else {
		depthArg = c.Num(float64(depth))
	}
	options := c.Obj()
	options.Set("depth", depthArg)
	options.SetBool("colors", false)
	options.SetBool("showHidden", st.opts.showHidden)
	options.SetBool("sorted", st.opts.sorted)
	options.SetBool("customInspect", st.opts.customInspect)
	options.SetBool("getters", st.opts.getters)
	if st.opts.breakLength > 0 {
		options.SetNum("breakLength", float64(st.opts.breakLength))
	} else {
		options.Set("breakLength", c.Num(math.Inf(1)))
	}
	if st.opts.maxArrayLength > 0 {
		options.SetNum("maxArrayLength", float64(st.opts.maxArrayLength))
	} else {
		options.Set("maxArrayLength", c.Null())
	}
	if st.opts.maxStringLength > 0 {
		options.SetNum("maxStringLength", float64(st.opts.maxStringLength))
	} else {
		options.Set("maxStringLength", c.Null())
	}
	options.SetNum("compact", 3)
	options.SetNum("indentationLvl", float64(indent))
	// Colors are never enabled, so stylize is the identity.
	options.SetFunc("stylize", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return c.Str(""), nil
		}
		return c.Str(args[0].Str()), nil
	})
	call := []js.Value{v, depthArg, options.V()}
	if customInspectFn.Valid() {
		call = append(call, customInspectFn)
	}
	r, err := customRender.Object().Call(js.Object{}, call...)
	if err != nil || !r.IsObject() {
		return "", false
	}
	if !r.Object().Get("hit").Bool() {
		return "", false
	}
	return r.Object().Get("text").Str(), true
}

func (st *inspectState) truncateString(s string) string {
	if st.opts.maxStringLength <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= st.opts.maxStringLength {
		return s
	}
	n := st.opts.maxStringLength
	missing := len(runes) - n
	return string(runes[:n]) + "... " + strconv.Itoa(missing) + " more character" + plural(missing)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func jsToString(v js.Value) string {
	if !v.Valid() {
		return "undefined"
	}
	// String(-0) is "0" but Node's formatter and inspector spell it "-0".
	if v.IsNumber() {
		f := v.Float()
		if f == 0 && math.Signbit(f) {
			return "-0"
		}
	}
	fn := v.Ctx().Global().Get("String")
	if fn.IsFunction() {
		if r, err := fn.Object().Call(js.Object{}, v); err == nil {
			return r.Str()
		}
	}
	return v.Str()
}
