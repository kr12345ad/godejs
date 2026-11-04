package env

import "regexp"

// V8 native syntax (for example %PrepareFunctionForOptimization) is accepted
// only by V8. Node's own tests use it inside literal eval calls to request
// optimization; JavaScriptCore cannot parse it. Strip those calls so the
// surrounding behavior tests still run. Ordinary eval is untouched.
var v8IntrinsicEval = regexp.MustCompile(`eval\(\s*['"]%[^'"]*['"]\s*\)`)

func stripV8IntrinsicEvals(src string) string {
	return v8IntrinsicEval.ReplaceAllString(src, "void 0")
}
