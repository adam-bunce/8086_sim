package main

import "fmt"

type Memory struct {
	Bytes []uint8
}

func (m *Memory) ReadMemory(address *int) uint8 {
	return m.Bytes[*address]
}

var MemoryValues = Memory{make([]uint8, 1024*1024)}

type CpuFlag int

const (
	SignFlag CpuFlag = iota
	ZeroFlag
)

func (c CpuFlag) String() string {
	if c == SignFlag {
		return "signFlag"
	} else {
		return "zeroFlag"
	}
}

type CpuFlags map[CpuFlag]bool

var CpuFlagValues = CpuFlags{
	SignFlag: false, // if the last op has a negative result signed is true
	ZeroFlag: false, // if the last op resulted in a value of 0 this is true
}

type Registers map[Register][]uint8

// RegisterValues hold the actual values in the registers
var RegisterValues = Registers{
	Register_a: []uint8{0b0, 0b0},
	Register_b: []uint8{0b0, 0b0},
	Register_c: []uint8{0b0, 0b0},
	Register_d: []uint8{0b0, 0b0},

	Register_sp: []uint8{0b0, 0b0},
	Register_bp: []uint8{0b0, 0b0},
	Register_si: []uint8{0b0, 0b0},
	Register_di: []uint8{0b0, 0b0},

	Register_ip: []uint8{0b0, 0b0},
}

func (r Registers) String() string {
	res := ""
	// it's a map so not ordered :x
	registersToPrint := len(RegisterValues)
	count := 1
	for count <= registersToPrint {
		memoryPlace := RegisterValues[Register(count)]
		res += fmt.Sprintf("%s\t%08b\t%v\t\n", Register(count), RegisterValues[Register(count)], ReadU16(memoryPlace, 0))
		count += 1
	}
	return res
}
