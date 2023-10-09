package main

import (
	"strconv"
)

var (
	D   = InstructionBits{Usage: Bits_D, BitCount: 1}
	W   = InstructionBits{Usage: Bits_W, BitCount: 1}
	S   = InstructionBits{Usage: Bits_S, BitCount: 1}
	RM  = InstructionBits{Usage: Bits_RM, BitCount: 3}
	MOD = InstructionBits{Usage: Bits_MOD, BitCount: 2}
	REG = InstructionBits{Usage: Bits_REG, BitCount: 3}

	DATA      = InstructionBits{Usage: Bits_Data, BitCount: 0}
	DATA_IF_W = InstructionBits{Usage: Bits_Data_If_W, BitCount: 0}

	DISP_LO = InstructionBits{Usage: Bits_Disp, BitCount: 0}
	DISP_HI = InstructionBits{Usage: Bits_Disp, BitCount: 0}

	ADDR_LO = InstructionBits{Usage: Bits_Disp, BitCount: 0}
	ADDR_HI = InstructionBits{Usage: Bits_Disp, BitCount: 0}

	IS_JUMP = InstructionBits{Usage: Bits_IsJump, BitCount: 0}
)

func ImpRm(rm uint8) InstructionBits {
	return InstructionBits{
		Usage:       Bits_RM,
		BitCount:    0,
		Value:       rm,
		HasValueSet: true,
	}
}
func ImpD(d uint8) InstructionBits {
	return InstructionBits{
		Usage:       Bits_D,
		BitCount:    0,
		Value:       d,
		HasValueSet: true,
	}
}

func ImpMod(mod uint8) InstructionBits {
	return InstructionBits{
		Usage:       Bits_MOD,
		BitCount:    0,
		Value:       mod,
		HasValueSet: true,
	}
}

func ImpReg(reg uint8) InstructionBits {
	return InstructionBits{
		Usage:       Bits_REG,
		BitCount:    0,
		Value:       reg,
		HasValueSet: true,
	}
}

func L(literalBits string) InstructionBits {
	value, _ := strconv.ParseUint(literalBits, 2, 8)
	return InstructionBits{
		Usage:    Bits_Literal,
		BitCount: uint8(len(literalBits)),
		Value:    uint8(value),
	}
}

var instTable = []InstructionEncoding{
	{Op_mov, []InstructionBits{L("100010"), D, W, MOD, REG, RM, DISP_LO, DISP_HI}},                              // Register/memory to/from register
	{Op_mov, []InstructionBits{L("1100011"), W, MOD, L("000"), RM, DATA, DATA_IF_W, ImpD(0)}},                   // Immediate to register/memory This somehow matches the 2nd add instruction (didn't break when false initially and it got reset to 000 so it ended up being true)
	{Op_mov, []InstructionBits{L("1011"), W, REG, DATA, DATA_IF_W, ImpD(1)}},                                    //Immediate to register, because source is always reg, d(swap src/dest) here is implicit
	{Op_mov, []InstructionBits{L("1010000"), W, ADDR_LO, ADDR_HI, ImpRm(0b110), ImpD(1), ImpMod(0), ImpReg(0)}}, // Memory to accumulator
	{Op_mov, []InstructionBits{L("1010001"), W, ADDR_LO, ADDR_HI, ImpRm(0b110), ImpD(0), ImpMod(0), ImpReg(0)}}, // Accumulator to memory
	// MISSING Register/memory to segment register
	// MISSING Segment register to register/memory

	{Op_add, []InstructionBits{L("000000"), D, W, MOD, REG, RM}},                       // Reg/memory with register to either
	{Op_add, []InstructionBits{L("100000"), S, W, MOD, L("000"), RM, DATA, DATA_IF_W}}, // Immediate to register/memory
	{Op_add, []InstructionBits{L("0000010"), W, DATA, DATA_IF_W, ImpReg(0), ImpD(1)}},  // Immediate to accumulator

	{Op_sub, []InstructionBits{L("001010"), D, W, MOD, REG, RM}},                       // Reg/memory with register to either
	{Op_sub, []InstructionBits{L("100000"), S, W, MOD, L("101"), RM, DATA, DATA_IF_W}}, // Immediate to register/memory
	{Op_sub, []InstructionBits{L("0010110"), W, DATA, DATA_IF_W, ImpReg(0), ImpD(1)}},  // Immediate to accumulator

	{Op_cmp, []InstructionBits{L("001110"), D, W, MOD, REG, RM}},                       // Reg/memory with register to either
	{Op_cmp, []InstructionBits{L("100000"), S, W, MOD, L("111"), RM, DATA, DATA_IF_W}}, // Immediate to register/memory
	{Op_cmp, []InstructionBits{L("0011110"), W, DATA, DATA_IF_W, ImpReg(0), ImpD(1)}},  // Immediate to accumulator

	// hops
	{Op_je, []InstructionBits{L("01110100"), DATA, IS_JUMP}},
	{Op_jl, []InstructionBits{L("01111100"), DATA, IS_JUMP}},
	{Op_jle, []InstructionBits{L("01111110"), DATA, IS_JUMP}},
	{Op_jb, []InstructionBits{L("01110010"), DATA, IS_JUMP}},
	{Op_jbe, []InstructionBits{L("01110110"), DATA, IS_JUMP}},
	{Op_js, []InstructionBits{L("01111000"), DATA, IS_JUMP}},
	{Op_jne, []InstructionBits{L("01110101"), DATA, IS_JUMP}},
	{Op_jnl, []InstructionBits{L("01111101"), DATA, IS_JUMP}},
	{Op_jg, []InstructionBits{L("01111111"), DATA, IS_JUMP}},
	{Op_ja, []InstructionBits{L("01110111"), DATA, IS_JUMP}},
	{Op_jns, []InstructionBits{L("01111001"), DATA, IS_JUMP}},
}
