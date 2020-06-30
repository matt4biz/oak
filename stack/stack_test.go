package stack

import (
	"reflect"
	"testing"
)

func TestMachinePushPop(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(3))
	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{1, 2, 3}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}

func TestMachineTop(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))

	var top float64

	if x := m.Top(); x != nil {
		if xf, ok := x.V.(float64); ok {
			top = xf
		}
	}

	want := 1.0

	if !reflect.DeepEqual(top, want) {
		t.Errorf("wanted %v, got %v", want, top)
	}
}

func TestMachineEmpty(t *testing.T) {
	m := &Machine{}

	if x := m.Top(); x != nil {
		t.Errorf("expected nil top-of-stack")
	}

	if x := m.Pop(); x != nil {
		t.Errorf("expected nil top-of-stack")
	}

	if x := m.PopX(); x != nil {
		t.Errorf("expected nil top-of-stack")
	}

	if x := m.Last(); x != nil {
		t.Errorf("expected nil top-of-stack")
	}
}

func TestMachinePopXLast(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))

	var top, last float64

	if x := m.PopX(); x != nil {
		if xf, ok := x.V.(float64); ok {
			top = xf
		}
	}

	wantTop := 1.0

	if !reflect.DeepEqual(top, wantTop) {
		t.Errorf("wanted %v, got %v", wantTop, top)
	}

	if x := m.Last(); x != nil {
		if xf, ok := x.V.(float64); ok {
			last = xf
		}
	}

	wantLast := 1.0

	if !reflect.DeepEqual(last, wantLast) {
		t.Errorf("wanted %v, got %v", wantLast, last)
	}

}

func TestMachinePopXLastSingle(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(1))

	var top, last float64

	if x := m.PopX(); x != nil {
		if xf, ok := x.V.(float64); ok {
			top = xf
		}
	}

	wantTop := 1.0

	if !reflect.DeepEqual(top, wantTop) {
		t.Errorf("wanted %v, got %v", wantTop, top)
	}

	if x := m.Last(); x != nil {
		if xf, ok := x.V.(float64); ok {
			last = xf
		}
	}

	wantLast := 1.0

	if !reflect.DeepEqual(last, wantLast) {
		t.Errorf("wanted %v, got %v", wantLast, last)
	}

}

func TestMachineDup(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))
	_ = m.Dup()

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{1, 1, 2}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}

func TestMachineDupEmpty(t *testing.T) {
	m := &Machine{}

	_ = m.Dup()

	if x := m.PopX(); x != nil {
		t.Errorf("expected nil top-of-stack")
	}
}

func TestMachineDup2(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))
	_ = m.Dup2()

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{1, 2, 1, 2}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}

func TestMachineRoll(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(3))
	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))
	m.Roll()

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{2, 3, 1}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}

func TestMachineRoll2(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))
	m.Roll()

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{2, 1}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}

func TestMachineSwap(t *testing.T) {
	m := &Machine{}

	m.Push(m.makeVal(3))
	m.Push(m.makeVal(2))
	m.Push(m.makeVal(1))
	_ = m.Swap()

	var stack []float64

	for x := m.Pop(); x != nil; x = m.Pop() {
		if xf, ok := x.V.(float64); ok {
			stack = append(stack, xf)
		}
	}

	want := []float64{2, 1, 3}

	if !reflect.DeepEqual(stack, want) {
		t.Errorf("wanted %v, got %v", want, stack)
	}
}
