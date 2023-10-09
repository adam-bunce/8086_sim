package main

import (
	"fmt"
)

func WriteU16(memory []uint8, position, value uint16) {
	// in little endian the least significant byte is stored at the lowest address
	memory[position] = uint8(value)
	memory[position+1] = uint8(value >> 8)
}

func ReadU16(memory []uint8, position uint16) uint16 {
	return uint16(memory[position]) | uint16(memory[position+1])<<8
}

func WriteU8(memory []uint8, position, value uint16) {
	memory[position] = uint8(value)
}

func ReadU8(memory []uint8, position uint16) uint8 {
	return memory[position]
}

func Write(memory []uint8, position, value uint16, isWide bool) {
	if isWide {
		WriteU16(memory, position, value)

	} else {
		WriteU8(memory, position, value)
	}
}

func ParseOperand(operand InstructionOperand) ([]uint8, uint16, uint16) {
	switch operand.Type {
	case Operand_Immediate:
		return nil, 0, uint16(uint(operand.Immediate.Value))
	case Operand_Register:
		var registerValue uint16
		if operand.Register.Length == 2 {
			// whole register
			registerValue = ReadU16(RegisterValues[operand.Register.RegisterIndex], uint16(operand.Register.ByteOffset))
		} else {
			// partial register read 1 byte from offset
			registerValue = uint16(ReadU8(RegisterValues[operand.Register.RegisterIndex], uint16(operand.Register.ByteOffset)))
		}
		return RegisterValues[operand.Register.RegisterIndex], uint16(operand.Register.ByteOffset), registerValue
	case Operand_Memory:
		loc := operand.EffectiveAddress.CalculateLocation()
		value := ReadU16(MemoryValues.Bytes, loc)
		return MemoryValues.Bytes, loc, value
	default:
		return nil, 0, 0
	}
}

// UpdateFlags updates flags based no the result of a sub/cmp operation
func UpdateFlags(result uint16) {
	if result == 0 {
		CpuFlagValues[ZeroFlag] = true
	} else {
		CpuFlagValues[ZeroFlag] = false
	}
	if result < 0 {
		CpuFlagValues[SignFlag] = true
	} else {
		CpuFlagValues[SignFlag] = false
	}
}

func HandleJump(jumpDistance uint16, flag bool) {
	var calcJumpDistance int

	// need to trim to 8bits for when i flip the bits
	trimmedJumpDistance := uint8(jumpDistance)

	// do two's complement if its negative
	if trimmedJumpDistance&(1<<7) != 0 {
		calcJumpDistance = int(^trimmedJumpDistance + 0b1)
		calcJumpDistance *= -1
	} else {
		calcJumpDistance = int(jumpDistance)
	}

	ipValue := ReadU16(RegisterValues[Register_ip], 0)

	if flag {
		newPosition := int(ipValue) + calcJumpDistance + 2
		WriteU16(RegisterValues[Register_ip], 0, uint16(newPosition))
	} else {
		WriteU16(RegisterValues[Register_ip], 0, ipValue+2) // jumps are 2 bytes
	}
}

func Simulate(instruction Instruction, doPrint bool) {
	if doPrint {
		fmt.Println(instruction)
	}

	dest, destPos, destValue := ParseOperand(instruction.InstructionOperands[0])
	_, _, srcValue := ParseOperand(instruction.InstructionOperands[1])
	isWide := instruction.Flags[Wide]

	switch instruction.Op {
	case Op_mov:
		Write(dest, destPos, srcValue, isWide)
	case Op_add:
		Write(dest, destPos, srcValue+destValue, isWide)
		UpdateFlags(destValue + srcValue)
	case Op_sub:
		Write(dest, destPos, destValue-srcValue, isWide)
		UpdateFlags(destValue - srcValue)
	case Op_cmp:
		UpdateFlags(destValue - srcValue)
	case Op_jne:
		HandleJump(srcValue, !CpuFlagValues[ZeroFlag])
		return
	case Op_je:
		HandleJump(srcValue, CpuFlagValues[ZeroFlag])
		return
	case Op_jl, Op_js:
		HandleJump(srcValue, CpuFlagValues[SignFlag])
		return
	case Op_jns:
		HandleJump(srcValue, !CpuFlagValues[SignFlag])
		return
	case Op_jnl, Op_jg:
		HandleJump(srcValue, !CpuFlagValues[SignFlag])
		return
	case Op_jle:
		HandleJump(srcValue, CpuFlagValues[ZeroFlag] || CpuFlagValues[SignFlag])
		return
	case Op_ja:
		HandleJump(srcValue, !CpuFlagValues[SignFlag] && !CpuFlagValues[ZeroFlag])
		return
	case Op_jbe:
		HandleJump(srcValue, CpuFlagValues[SignFlag] || CpuFlagValues[ZeroFlag])
		return
	case Op_jb:
		HandleJump(srcValue, CpuFlagValues[SignFlag] && !CpuFlagValues[ZeroFlag])
		return
	}
	// move IP
	WriteU16(RegisterValues[Register_ip], 0, ReadU16(RegisterValues[Register_ip], 0)+uint16(instruction.Size))
}
