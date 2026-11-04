package util

import (
	"strings"

	"godejs/internal/js"
)

// Format implements util.format.
func Format(args []js.Value) string {
	return FormatWith(args, false)
}

// FormatWith implements util.formatWithOptions' formatting half. Only the
// numericSeparator option changes the text util.format produces.
func FormatWith(args []js.Value, numericSeparator bool) string {
	if len(args) == 0 {
		return ""
	}
	if first := args[0]; first.IsString() {
		return formatString(first.Str(), args[1:], numericSeparator)
	}
	parts := make([]string, 0, len(args))
	for _, a := range args {
		parts = append(parts, formatTopLevel(a, numericSeparator))
	}
	return strings.Join(parts, " ")
}

func formatTopLevel(v js.Value, numericSeparator bool) string {
	if v.IsString() {
		return v.Str()
	}
	if numericSeparator && v.IsNumber() {
		return formatNumberWith(v, true)
	}
	return Inspect(v)
}

func formatString(format string, args []js.Value, numericSeparator bool) string {
	var b strings.Builder
	rest := args
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		spec := format[i+1]
		switch spec {
		case '%':
			b.WriteByte('%')
			i++
		case 'c':
			// CSS styling is ignored, but the argument is consumed.
			i++
			if len(rest) > 0 {
				rest = rest[1:]
			}
		case 's', 'd', 'i', 'f', 'j', 'o', 'O':
			if len(rest) == 0 {
				b.WriteByte('%')
				b.WriteByte(spec)
				i++
				continue
			}
			arg := rest[0]
			rest = rest[1:]
			b.WriteString(specValue(spec, arg, numericSeparator))
			i++
		default:
			b.WriteByte('%')
		}
	}
	for _, a := range rest {
		b.WriteByte(' ')
		b.WriteString(formatTopLevel(a, numericSeparator))
	}
	return b.String()
}

func specValue(spec byte, v js.Value, numericSeparator bool) string {
	switch spec {
	case 's':
		if v.IsString() {
			return v.Str()
		}
		if v.IsNumber() || v.Tag() == "BigInt" {
			return formatNumberWith(v, numericSeparator)
		}
		if v.IsObject() || v.IsFunction() {
			return Inspect(v)
		}
		return jsToString(v)
	case 'd':
		return numberText(v, numericSeparator)
	case 'i':
		// Node formats a BigInt argument directly instead of running parseInt.
		if v.Tag() == "BigInt" {
			return formatNumberWith(v, numericSeparator)
		}
		// Node runs the global parseInt, whose ToString step turns the number
		// -0 into "0" while "-0.5" truncates to -0.
		if pi := v.Ctx().Global().Get("parseInt"); pi.IsFunction() {
			if r, err := pi.Object().Call(js.Object{}, v, v.Ctx().Num(10)); err == nil && r.IsNumber() {
				return formatNumberWith(r, numericSeparator)
			}
		}
		return numberText(v, numericSeparator)
	case 'f':
		if pf := v.Ctx().Global().Get("parseFloat"); pf.IsFunction() {
			if r, err := pf.Object().Call(js.Object{}, v); err == nil && r.IsNumber() {
				return formatNumberWith(r, numericSeparator)
			}
		}
		return jsToString(v.Ctx().Num(v.Float()))
	case 'j':
		// Symbols have no JSON form; Node renders them (and undefined and
		// functions) as the literal text "undefined".
		if v.IsUndefined() || v.IsFunction() || (!v.IsObject() && v.Tag() == "Symbol") {
			return "undefined"
		}
		s, err := v.JSON(0)
		if err != nil {
			return "[Circular]"
		}
		return s
	case 'o':
		return inspectWith(v, inspectOptions{depth: 4, showHidden: true, breakLength: 80, compact: true, customInspect: true, numericSeparator: numericSeparator})
	case 'O':
		opts := defaultInspectOptions
		opts.numericSeparator = numericSeparator
		return inspectWith(v, opts)
	}
	return ""
}

// formatNumberWith renders a number the way Node's formatNumber does: plain
// String() output unless numericSeparator asks for grouping. Exponential
// forms and non-finite values are left untouched.
func formatNumberWith(v js.Value, numericSeparator bool) string {
	if v.Tag() == "BigInt" {
		s := jsToString(v)
		if numericSeparator {
			neg := strings.HasPrefix(s, "-")
			digits := strings.TrimPrefix(s, "-")
			digits = groupInteger(digits)
			if neg {
				digits = "-" + digits
			}
			s = digits
		}
		return s + "n"
	}
	s := formatNumber(v)
	if !numericSeparator || strings.ContainsAny(s, "eE") {
		return s
	}
	if s == "NaN" || s == "Infinity" || s == "-Infinity" {
		return s
	}
	neg := strings.HasPrefix(s, "-")
	body := s
	if neg {
		body = body[1:]
	}
	intPart, frac, hasDot := strings.Cut(body, ".")
	intPart = groupInteger(intPart)
	if hasDot {
		frac = groupFraction(frac)
		body = intPart + "." + frac
	} else {
		body = intPart
	}
	if neg {
		return "-" + body
	}
	return body
}

// groupInteger inserts an underscore every three digits from the right, never
// before the leading one to three digits.
func groupInteger(s string) string {
	n := len(s)
	if n < 4 {
		return s
	}
	var b strings.Builder
	first := n % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < n; i += 3 {
		b.WriteByte('_')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// groupFraction inserts an underscore every three digits from the left; the
// final one to three digits stay unseparated.
func groupFraction(s string) string {
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i += 3 {
		if i > 0 {
			b.WriteByte('_')
		}
		end := i + 3
		if end > len(s) {
			end = len(s)
		}
		b.WriteString(s[i:end])
	}
	return b.String()
}

func numberText(v js.Value, numericSeparator bool) string {
	if v.IsNumber() {
		return formatNumberWith(v, numericSeparator)
	}
	if v.Tag() == "BigInt" {
		// Node's formatBigIntNoColor appends the BigInt suffix.
		return formatNumberWith(v, numericSeparator)
	}
	if v.IsString() || v.IsBool() || v.IsNull() {
		// Number() preserves the sign of zero for strings like '-0.000'.
		num := v.Ctx().Global().Get("Number")
		if num.IsFunction() {
			if r, err := num.Object().Call(js.Object{}, v); err == nil && r.IsNumber() {
				return formatNumberWith(r, numericSeparator)
			}
		}
		return jsToString(v.Ctx().Num(v.Float()))
	}
	return "NaN"
}
