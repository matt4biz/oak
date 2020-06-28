package stack

func Nop(m *Machine) error {
	return nil
}

// Last returns the last top-of-stack value that
// was popped with PopX for use in a calculation.
func (m *Machine) Last() *Value {
	return m.x
}

func (m *Machine) Top() *Value {
	if l := len(m.stack); l > 0 {
		return m.stack[l-1]
	}

	return nil
}

// Pop just removes and returns the top of stack.
func (m *Machine) Pop() *Value {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		return i
	}
}

// PopX is used when we're going to push back some
// result involving the TOS, so Last() can return it.
func (m *Machine) PopX() *Value {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		m.x = i
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		m.x = i
		return i
	}
}

func (m *Machine) Push(v Value) {
	m.stack = append(m.stack, &v)
}

func (m *Machine) Swap() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack[0:l-2], m.stack[l-1], m.stack[l-2])
	}
}

// Dup takes the top-of-stack (x) and duplicates
// it on top of the stack.
func (m *Machine) Dup() {
	if l := len(m.stack); l > 0 {
		m.stack = append(m.stack, m.stack[l-1])
	}
}

// Dup2 takes {y,x} and duplicates those two values
// in order on top of the stack.
func (m *Machine) Dup2() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack, m.stack[l-2], m.stack[l-1])
	}
}

// Roll moves the top-of-stack item onto the bottom,
// exposing the next item (y) as the top.
func (m *Machine) Roll() {
	switch l := len(m.stack); l {
	case 0, 1:
		// nothing to do

	case 2:
		m.Swap()

	default:
		tmp := []*Value{m.stack[l-1]}

		m.stack = append(tmp, m.stack[0:l-1]...)
	}
}

func (m *Machine) SetFixed(d int) {
	m.disp = fixed
	m.digits = d
}

func (m *Machine) SetScientific(d int) {
	m.disp = scientific
	m.digits = d
}

func (m *Machine) SetEngineering(d int) {
	m.disp = engineering
	m.digits = d
}

func (m *Machine) SetRadians() {
	m.mode = radians
}
