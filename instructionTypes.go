package main

import (
	"fmt"
	"strconv"
)

type InstructionBitsUsage int

const (
	Bits_Literal InstructionBitsUsage = iota

	Bits_D
	Bits_S
	Bits_W
	Bits_MOD
	Bits_REG
	Bits_RM
	Bits_Disp
	Bits_Data
	Bits_Data_If_W

	Bits_IsJump
)

// InstructionBits are some part of the instruction, could be mod/reg/rm/whatever
type InstructionBits struct {
	Usage    InstructionBitsUsage
	BitCount uint8

	HasValueSet bool
	Value       uint8 //  forcibly set value of Instruction bits during decode rather than read bits
}

type OperationType int

const (
	Op_mov OperationType = iota
	Op_add
	Op_sub
	Op_cmp
	Op_je
	Op_jl
	Op_jle
	Op_jb
	Op_jbe
	Op_js
	Op_jne
	Op_jnl
	Op_jg
	Op_ja
	Op_jns
)

var opTypeToString = map[OperationType]string{
	Op_mov: "mov",
	Op_add: "add",
	Op_sub: "sub",
	Op_cmp: "cmp",

	Op_je:  "je",
	Op_jl:  "jl",
	Op_jle: "jle",
	Op_jb:  "jb",
	Op_jbe: "jbe",
	Op_js:  "js",
	Op_jne: "jne",
	Op_jnl: "jnl",
	Op_jg:  "jg",
	Op_ja:  "ja",
	Op_jns: "jns",
}

type InstructionEncoding struct {
	Op   OperationType
	Bits []InstructionBits
}

type OperandType int

const (
	Operand_None OperandType = iota
	Operand_Register
	Operand_Memory
	Operand_Immediate
)

type Register int

const (
	Register_none Register = iota // why do i have this

	Register_a
	Register_b
	Register_c
	Register_d
	Register_sp
	Register_bp
	Register_si
	Register_di

	Register_ip // what
)

func (r Register) String() string {
	return []string{"none", "a", "b", "c", "d", "sp", "bp", "si", "di", "ip"}[r]
}

// Flag - set during decode stage, different from flags involved in simulation
type Flag int

const (
	Wide Flag = iota
	IsJump
)

// InstructionOperand represents the possible operands that can be passed to an instruction
// can be effective address [bp + 5], register access ax, or immediate -12
type InstructionOperand struct {
	Type OperandType

	Register         RegisterAccess   // Operand_Register
	Immediate        Immediate        // Operand_Immediate
	EffectiveAddress EffectiveAddress // Operand_Memory
}

func (iop InstructionOperand) String() string {
	switch iop.Type {
	case Operand_Immediate:
		return strconv.Itoa(iop.Immediate.Value)
	case Operand_Register:
		return iop.Register.String()
	case Operand_Memory:
		return iop.EffectiveAddress.String()
	}
	return "unexpected iop got " + strconv.Itoa(int(iop.Type))
}

type EffectiveAddress struct {
	EffectiveAddressExpression EffectiveAddressFieldEncoding // bx + si, bx + di, dp + di... etc whatever
	Displacement               int
	Size                       int // byte or word
}

func (e EffectiveAddress) CalculateLocation() uint16 {
	var location uint16
	switch e.EffectiveAddressExpression {
	case EffectiveAddress_bx_si:
		location = ReadU16(RegisterValues[Register_b], 0) + ReadU16(RegisterValues[Register_si], 0)
	case EffectiveAddress_bx_di:
		location = ReadU16(RegisterValues[Register_b], 0) + ReadU16(RegisterValues[Register_di], 0)
	case EffectiveAddress_bp_si:
		location = ReadU16(RegisterValues[Register_bp], 0) + ReadU16(RegisterValues[Register_si], 0)
	case EffectiveAddress_bp_di:
		location = ReadU16(RegisterValues[Register_bp], 0) + ReadU16(RegisterValues[Register_di], 0)
	case EffectiveAddress_si:
		location = ReadU16(RegisterValues[Register_si], 0)
	case EffectiveAddress_di:
		location = ReadU16(RegisterValues[Register_di], 0)
	case EffectiveAddress_bp:
		location = ReadU16(RegisterValues[Register_bp], 0)
	case EffectiveAddress_bx:
		location = ReadU16(RegisterValues[Register_b], 0)
	}
	location += uint16(e.Displacement)
	return location
}

var EffectiveAddressFieldEncodingToString = map[EffectiveAddressFieldEncoding]string{
	EffectiveAddress_bx_si: "bx + si",
	EffectiveAddress_bx_di: "bx + di",
	EffectiveAddress_bp_si: "bp + si",
	EffectiveAddress_bp_di: "bp + di",
	EffectiveAddress_si:    "si",
	EffectiveAddress_di:    "di",
	EffectiveAddress_bp:    "bp",
	EffectiveAddress_bx:    "bx",

	EffectiveAddress_Direct_Address: "DIRECT ADDRESS",
}

func (e EffectiveAddress) String() string {
	res := "["
	if e.EffectiveAddressExpression == EffectiveAddress_Direct_Address {
		res += strconv.Itoa(e.Displacement) + "]"
		return res
	}

	res += EffectiveAddressFieldEncodingToString[e.EffectiveAddressExpression]
	if e.Displacement != 0 {
		res += " + " + strconv.Itoa(e.Displacement)
	}
	res += "]"
	return res
}

type EffectiveAddressFieldEncoding int

const (
	// reg maps into these directly can use EffectiveAddressFieldEncoding(register_bits)
	EffectiveAddress_bx_si EffectiveAddressFieldEncoding = iota
	EffectiveAddress_bx_di
	EffectiveAddress_bp_si
	EffectiveAddress_bp_di
	EffectiveAddress_si
	EffectiveAddress_di
	EffectiveAddress_bp
	EffectiveAddress_bx

	EffectiveAddress_Direct_Address
)

type Immediate struct {
	Value int
}

type RegisterAccess struct {
	RegisterIndex Register
	ByteOffset    uint
	Length        uint
}

func (r RegisterAccess) String() string {
	var registerNames = [][]string{
		{"", "", ""},
		{"al", "ah", "ax"},
		{"bl", "bh", "bx"},
		{"cl", "ch", "cx"},
		{"dl", "dh", "dx"},
		{"sp", "sp", "sp"},
		{"bp", "bp", "bp"},
		{"si", "si", "si"},
		{"di", "di", "di"},
	}

	if r.Length == 2 {
		return registerNames[r.RegisterIndex][r.Length]
	} else {
		return registerNames[r.RegisterIndex][r.ByteOffset]
	}
}

type Instruction struct {
	Address uint32
	Size    uint32
	Bytes   []byte

	Op    OperationType
	Flags map[Flag]bool

	InstructionOperands [2]InstructionOperand
}

func (i Instruction) String() string {
	var sizePrefix string
	if i.InstructionOperands[0].Type != Operand_Register &&
		i.InstructionOperands[1].Type != Operand_Register {
		if i.Flags[Wide] {
			sizePrefix = "word"
		} else {
			sizePrefix = "byte"
		}
	}
	if i.Flags[IsJump] {
		return fmt.Sprintf("%s %s", opTypeToString[i.Op], i.InstructionOperands[1])
	}

	return fmt.Sprintf("%s %s %s, %s", opTypeToString[i.Op], sizePrefix, i.InstructionOperands[0], i.InstructionOperands[1])
}
