package main

import (
	"fmt"
	"maps"
	"slices"
)

var totalCycles = 0
var tookJump = false

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

func HandleJump(jumpDistance uint16, flag bool, instSize uint32) {
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
		tookJump = true
		newPosition := int(ipValue) + calcJumpDistance + int(instSize)
		WriteU16(RegisterValues[Register_ip], 0, uint16(newPosition))
	} else {
		tookJump = false
		WriteU16(RegisterValues[Register_ip], 0, ipValue+uint16(instSize)) // jumps are 2 bytes, call is 3
	}
}

func PushValueToStack(value uint16) {
	spValue := ReadU16(RegisterValues[Register_sp], 0)     // get current stack position
	Write(MemoryValues.Bytes, spValue-2, value, true)      // write new value to stack
	Write(RegisterValues[Register_sp], 0, spValue-2, true) // update stack pointer
}

func PopValueFromStack(dest []uint8, destPos uint16) {
	spValue := ReadU16(RegisterValues[Register_sp], 0) // get current stack position
	stackValue := ReadU16(MemoryValues.Bytes, spValue)
	if spValue == 40_000 {
		panic("attempt to pop from empty stack")
	}
	Write(MemoryValues.Bytes, spValue, 0, true)            // clear values on stack (set to 0) (lowk don't need)
	Write(RegisterValues[Register_sp], 0, spValue+2, true) // update stack pointer
	Write(dest, destPos, stackValue, true)                 // put value from sp into dest
}

func HandlePrint(instruction Instruction, showEffect []bool, initalIp uint16) {
	_, _, destValue := ParseOperand(instruction.InstructionOperands[0])
	_, _, srcValue := ParseOperand(instruction.InstructionOperands[1])

	currFlags := CpuFlags{}
	maps.Copy(currFlags, CpuFlagValues)

	if showEffect[ShowInst] {
		fmt.Printf("%-30s", instruction)
	}

	if showEffect[ShowInstBytes] {
		fmt.Printf("%08b ", instruction.Bytes)
	}

	if showEffect[ShowCycles] {
		cycles := CalculateInstructionCycles(instruction, tookJump)
		totalCycles += cycles
		fmt.Printf("%-25s", fmt.Sprintf(" cycles: + %d = %d ", cycles, totalCycles))
	}

	if showEffect[ShowInst] {
		if instruction.InstructionOperands[0].Type == Operand_Register {
			// print register update if there is one
			fmt.Printf("%s", fmt.Sprintf("%s:%x->%x ", instruction.InstructionOperands[0].Register.String(), destValue, srcValue))
		}
		fmt.Printf("%s", fmt.Sprintf("ip:%x->%x ", initalIp, ReadU16(RegisterValues[Register_ip], 0)))

		// print flags
		isDiff := false
		flagInfo := "Flags: "
		for key := range currFlags {
			if currFlags[key] != CpuFlagValues[key] {
				isDiff = true
				flagInfo += fmt.Sprintf("%s->%v ", key, CpuFlagValues[key])
			}
		}
		if isDiff {
			fmt.Printf(flagInfo)
		}
	}
	if slices.Contains(showEffect, true) {
		fmt.Println()
	}
}

func Simulate(instruction Instruction, showEffect []bool) {
	initialIPVal := ReadU16(RegisterValues[Register_ip], 0)
	defer HandlePrint(instruction, showEffect, initialIPVal)

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
		HandleJump(srcValue, !CpuFlagValues[ZeroFlag], 2)
		return
	case Op_je:
		HandleJump(srcValue, CpuFlagValues[ZeroFlag], instruction.Size)
		return
	case Op_jl, Op_js:
		HandleJump(srcValue, CpuFlagValues[SignFlag], instruction.Size)
		return
	case Op_jns:
		HandleJump(srcValue, !CpuFlagValues[SignFlag], instruction.Size)
		return
	case Op_jnl, Op_jg:
		HandleJump(srcValue, !CpuFlagValues[SignFlag], instruction.Size)
		return
	case Op_jle:
		HandleJump(srcValue, CpuFlagValues[ZeroFlag] || CpuFlagValues[SignFlag], instruction.Size)
		return
	case Op_ja:
		HandleJump(srcValue, !CpuFlagValues[SignFlag] && !CpuFlagValues[ZeroFlag], instruction.Size)
		return
	case Op_jbe:
		HandleJump(srcValue, CpuFlagValues[SignFlag] || CpuFlagValues[ZeroFlag], instruction.Size)
		return
	case Op_jb:
		HandleJump(srcValue, CpuFlagValues[SignFlag] && !CpuFlagValues[ZeroFlag], instruction.Size)
		return
	case Op_jmp:
		HandleJump(srcValue, true, instruction.Size)
		return
	case Op_push:
		PushValueToStack(destValue)
	case Op_pop:
		PopValueFromStack(dest, destPos)
	case Op_call:
		PushValueToStack(ReadU16(RegisterValues[Register_ip], 0) + uint16(instruction.Size)) // end of this instruction not start
		HandleJump(srcValue, true, instruction.Size)
		return
	case Op_ret:
		PopValueFromStack(RegisterValues[Register_ip], 0)
		return
	default:
		panic(fmt.Sprintf("unimplemented instruction %v", instruction))
	}

	// move IP
	WriteU16(RegisterValues[Register_ip], 0, ReadU16(RegisterValues[Register_ip], 0)+uint16(instruction.Size))
}
