package errors

import "testing"

func TestCodeAndString(t *testing.T) {
	tests := []struct {
		e error
		c int
		s string
	}{{
		e: E("err"),
		c: 500,
		s: "Other error",
	}, {
		e: E(Unauthorized),
		c: 401,
		s: "Unauthorized",
	}, {
		e: E(NotFound),
		c: 404,
		s: "NotFound",
	}, {
		e: E(Duplicate),
		c: 409,
		s: "Duplicate",
	}, {
		e: E(Invalid),
		c: 400,
		s: "Invalid input",
	}}
	for _, tst := range tests {
		if e, ok := tst.e.(*Error); !ok {
			t.Error("Expected err to be of type *Error")
		} else if e.Code() != tst.c {
			t.Errorf("Expected code to be %d, got %d", tst.c, e.Code())
		}
	}
}

func TestOps(t *testing.T) {
	e1 := E(Op("op1"), "err").(*Error)
	compareOps(t, e1.Ops(), []Op{"op1"})

	e2 := E(Op("anotherOp"), e1).(*Error)
	compareOps(t, e2.Ops(), []Op{"op1", "anotherOp"})

	e3 := E(Op("newOp"), e2).(*Error)
	compareOps(t, e3.Ops(), []Op{"op1", "anotherOp", "newOp"})

	e4 := E(Op("op"), e2).(*Error)
	compareOps(t, e4.Ops(), []Op{"op1", "anotherOp", "op"})
}

func compareOps(t *testing.T, o1 []Op, o2 []Op) {
	t.Helper()
	if len(o1) != len(o2) {
		t.Error("Ops lengths must not be different.")
		return
	}
	for i, o := range o1 {
		if o2[i] != o {
			t.Errorf("Expected %s to be %s", o2[i], o)
		}
	}
}
