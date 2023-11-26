package main

import (
	"fmt"
)

func Decode(memory []byte, at *int) map[int]Instruction {
	memVal := Memory{Bytes: memory}
	var allInstructions = map[int]Instruction{}

	for *at < len(memory) {
		atBeforeTryingInstructions := *at
		for _, instruction := range instTable {
			instVal, err := TryDecode(memVal, *at, instruction)
			if err != nil {
				continue
			}

			// instruction was valid no need to test more
			allInstructions[*at] = instVal
			*at += int(instVal.Size) // confirm the read, move the address the length of the instruction to the next instruction
			break                    // found a match, we can exit the inner loop and try decoding from the next readable byte
		}

		// we failed to decode
		if *at == atBeforeTryingInstructions {
			fmt.Println(allInstructions)
			panic(fmt.Sprintf("[ERROR]: failed to decode: %08b\n", memory[*at:]))

		}

	}

	return allInstructions
}

// TryDecode attempts to decode one(1) instruction, and moves the at position forwards
func TryDecode(memory Memory, at int, possibleInstruction InstructionEncoding) (Instruction, error) {
	isValidInst := true
	decodedInst := Instruction{
		Address:             uint32(at),
		Size:                1, // length in bytes
		Op:                  possibleInstruction.Op,
		Flags:               map[Flag]bool{},
		InstructionOperands: [2]InstructionOperand{},
	}

	// bits in the instruction
	var bits = make(map[InstructionBitsUsage]byte)
	var has = make(map[InstructionBitsUsage]bool)

	testBits := InstructionBits{
		BitCount: 8,
		Value:    memory.ReadMemory(&at),
	}
	decodedInst.Bytes = append(decodedInst.Bytes, memory.ReadMemory(&at))
	at += 1 // don't move at until we actually read

	for _, pisBits := range possibleInstruction.Bits {
		if testBits.BitCount == 0 && at < len(memory.Bytes) && pisBits.BitCount != 0 {
			testBits.Value = memory.ReadMemory(&at)
			decodedInst.Bytes = append(decodedInst.Bytes, memory.ReadMemory(&at))
			at += 1 // don't move at until we actually read
			decodedInst.Size++
		}

		shiftDistance := 8 - pisBits.BitCount
		if pisBits.Usage == Bits_Literal {
			// valid instruction if all the literal bits match
			if pisBits.Value == testBits.Value>>shiftDistance {
				isValidInst = true
			} else {
				isValidInst = false
				break
			}
		} else {
			bits[pisBits.Usage] = testBits.Value >> shiftDistance

			if pisBits.HasValueSet {
				// overwrite b/c its implicit or something
				bits[pisBits.Usage] = pisBits.Value
			}

			has[pisBits.Usage] = true
		}

		testBits.BitCount -= pisBits.BitCount
		testBits.Value <<= pisBits.BitCount // remove used bits for readable section

	}

	if isValidInst {
		mod := bits[Bits_MOD] // memory mode information stored here
		reg := bits[Bits_REG] // how the effective address of the memory operand is to be calculated
		rm := bits[Bits_RM]
		d := bits[Bits_D] // instruction source is specified in REG (swap reg/rm)
		w := bits[Bits_W]

		// page 83
		hasDirectAddress := (rm == 0b110) && (mod == 0b00)
		hasDisplacement := (mod == 0b10) || (mod == 0b01 || hasDirectAddress)
		displacementIsW := (mod == 0b10) || hasDirectAddress
		dataisW := bits[Bits_S] != 1 && (w == 0b1)

		DispVal := ParseDataValue(&memory, &at, &decodedInst, hasDisplacement, displacementIsW)
		DataVal := ParseDataValue(&memory, &at, &decodedInst, has[Bits_Data], dataisW)

		source := &decodedInst.InstructionOperands[1]
		dest := &decodedInst.InstructionOperands[0]

		// Set Flags
		if w == 0b1 {
			// if w is set, dest/src must be wide
			decodedInst.Flags[Wide] = true
		}
		if has[Bits_IsJump] {
			decodedInst.Flags[IsJump] = true
		}

		// swap src/dest depending on d
		if d == 0b1 {
			// instruction SOURCE is specified in REG field
			source, dest = dest, source
		}

		if has[Bits_REG] {
			*source = GetRegisterOperand(reg, w)
		}

		if has[Bits_MOD] {
			if mod == 0b11 {
				*dest = GetRegisterOperand(rm, w)
			} else {
				// Effective Address Calculation
				(*dest).Type = Operand_Memory
				(*dest).EffectiveAddress.Displacement = int(DispVal)

				if w == 0b1 {
					(*dest).EffectiveAddress.Size = Word
				} else {
					(*dest).EffectiveAddress.Size = Byte
				}

				if mod == 0b00 && rm == 0b110 {
					(*dest).EffectiveAddress.EffectiveAddressExpression = EffectiveAddress_Direct_Address
				} else {
					(*dest).EffectiveAddress.EffectiveAddressExpression = EffectiveAddressFieldEncoding(rm)
				}
			}
		}
		if has[Bits_Data] {
			//always 2nd operand
			decodedInst.InstructionOperands[1].Type = Operand_Immediate
			decodedInst.InstructionOperands[1].Immediate.Value = int(DataVal)

		}
		return decodedInst, nil
	}

	return Instruction{}, fmt.Errorf("instruction didn't match. Invalid.")
}

func GetRegisterOperand(registerIndex uint8, w uint8) InstructionOperand {
	operand := InstructionOperand{
		Type:     Operand_Register,
		Register: RegisterAccess{},
	}

	registerFieldEncoding := [][2]RegisterAccess{
		// {bx, 0, 2}, means bx, start from bl (first bit of entire register) and go read 2 bytes (entire register)
		0: {{Register_a, 0, 1}, {Register_a, 0, 2}},  // al ax
		1: {{Register_c, 0, 1}, {Register_c, 0, 2}},  // cl cx
		2: {{Register_d, 0, 1}, {Register_d, 0, 2}},  // dl dx
		3: {{Register_b, 0, 1}, {Register_b, 0, 2}},  // bl bx
		4: {{Register_a, 1, 1}, {Register_sp, 0, 2}}, // ah sp
		5: {{Register_c, 1, 1}, {Register_bp, 0, 2}}, // ch bp
		6: {{Register_d, 1, 1}, {Register_si, 0, 2}}, // dh si
		7: {{Register_b, 1, 1}, {Register_di, 0, 2}}, // bh di
	}

	operand.Register = registerFieldEncoding[registerIndex][w]

	return operand
}

func ParseDataValue(memory *Memory, at *int, decodedInst *Instruction, exists, wide bool) uint16 {
	var res uint16
	if exists {
		if wide {
			// read 2 bytes
			b1 := memory.ReadMemory(at)
			decodedInst.Bytes = append(decodedInst.Bytes, b1)
			*at += 1
			decodedInst.Size++
			b2 := memory.ReadMemory(at)
			decodedInst.Bytes = append(decodedInst.Bytes, b2)
			decodedInst.Size++
			*at += 1
			res = uint16(b1) | uint16(b2)<<8
		} else {
			// read
			b1 := memory.ReadMemory(at)
			decodedInst.Bytes = append(decodedInst.Bytes, b1)
			*at += 1
			decodedInst.Size++
			res = uint16(b1)
		}
	}
	return res
}
