package uictl

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// Knob is a simple on/off toggle control.
type Knob interface {
	Read() bool
	On()
	Off()
	Toggle()
}

// Dial is a control that can read some value.
type Dial[N Number] interface {
	Read() N
}

// CappedDial is a Dial with a maximum cap value.
type CappedDial[N Number] interface {
	Dial[N]
	Cap() (num, max N)
}

// Levels is a control that can read multiple float32 levels.
type Levels[N Number] interface {
	Read() []N
}
