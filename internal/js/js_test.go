package js

import "testing"

func TestEvalAndHostFunction(t *testing.T) {
	c := New()
	defer c.Close()

	g := c.Global()
	g.SetNum("base", 40)
	g.SetFunc("add", func(this Object, args []Value) (Value, error) {
		sum := 0.0
		for _, a := range args {
			sum += a.Float()
		}
		return c.Num(sum), nil
	})

	v, err := c.Eval(`add(base, 2)`, "smoke.js")
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got, want := v.Str(), "42"; got != want {
		t.Fatalf("eval result = %q, want %q", got, want)
	}
}

func TestTypesAndContainers(t *testing.T) {
	c := New()
	defer c.Close()

	obj := c.Obj()
	obj.SetStr("name", "gode")
	obj.SetNum("n", 7)
	obj.SetFunc("double", func(this Object, args []Value) (Value, error) {
		return c.Num(args[0].Float() * 2), nil
	})
	c.Global().SetObj("o", obj)

	v, err := c.Eval(`o.name + ":" + o.double(o.n)`, "types.js")
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got, want := v.Str(), "gode:14"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	arr, err := c.Eval(`[1, 2, 3]`, "arr.js")
	if err != nil {
		t.Fatalf("eval array: %v", err)
	}
	if !arr.IsArray() || arr.Object().Length() != 3 || arr.Object().At(2).Int() != 3 {
		t.Fatalf("array decode failed: %#v", arr.Export())
	}
}

func TestBytesAndErrors(t *testing.T) {
	c := New()
	defer c.Close()

	b := c.Bytes([]byte{1, 2, 3})
	if got, ok := b.V().Bytes(); !ok || len(got) != 3 || got[2] != 3 {
		t.Fatalf("bytes roundtrip failed: %v %v", got, ok)
	}

	if _, err := c.Eval(`throw new TypeError("nope")`, "err.js"); err == nil {
		t.Fatal("expected an error")
	} else if jsErr, ok := err.(*Error); !ok || jsErr.Name != "TypeError" || jsErr.Message != "nope" {
		t.Fatalf("wrong error: %#v", err)
	}

	if _, err := c.Eval(`nope()`, "ref.js"); err == nil {
		t.Fatal("expected a ReferenceError")
	}
}

func TestPrototype(t *testing.T) {
	c := New()
	defer c.Close()

	proto := c.Obj()
	proto.SetStr("greet", "hi")
	o := c.Obj()
	o.SetProto(proto)
	c.Global().SetObj("msg", o)
	v, err := c.Eval(`msg.greet`, "proto.js")
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if v.Str() != "hi" {
		t.Fatalf("prototype lookup failed: %q", v.Str())
	}
}
