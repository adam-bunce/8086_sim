package main

// CalculateCycles calculates the cycles of an instruction given its baseClocks, tranfers and the memory operand
// if memory operand is nil then it's assume theres no EA calc to be done
func CalculateCycles(baseClocks, transfers int, ea *EffectiveAddress) int {
	var transferPenaltyCycles, eaCycles int
	cumCycles := 0
	cumCycles += baseClocks
	if ea != nil {
		eaCycles = calculateEACycles(ea)
		transferPenaltyCycles = calculateTransferCycles(ea, transfers)
	}
	cumCycles += transferPenaltyCycles
	cumCycles += eaCycles
	return cumCycles
}

func calculateTransferCycles(ea *EffectiveAddress, transfers int) int {
	if ea.Size == Word && ea.CalculateLocation()%2 != 0 {
		return 4 * transfers
	}
	return 0
}

func calculateEACycles(ea *EffectiveAddress) int {
	if ea == nil {
		return 0
	}

	cum := 0

	if ea.Displacement == 0 {
		// no displacement
		switch ea.EffectiveAddressExpression {
		case EffectiveAddress_bx, EffectiveAddress_bp, EffectiveAddress_si, EffectiveAddress_di:
			cum += 5
		case EffectiveAddress_bp_di, EffectiveAddress_bx_si:
			cum += 7
		case EffectiveAddress_bp_si, EffectiveAddress_bx_di:
			cum += 8
		}
	} else {
		switch ea.EffectiveAddressExpression {
		case EffectiveAddress_Direct_Address:
			cum += 6
		case EffectiveAddress_bx, EffectiveAddress_bp, EffectiveAddress_si, EffectiveAddress_di:
			cum += 9
		case EffectiveAddress_bp_di, EffectiveAddress_bx_si:
			cum += 11
		case EffectiveAddress_bp_si, EffectiveAddress_bx_di:
			cum += 12
		}
	}
	return cum
}

func instOperandIsType(operand InstructionOperand, operandType OperandType) bool {
	return operand.Type == operandType
}

func isAccumulatorUsed(operands ...InstructionOperand) bool {
	for _, operand := range operands {
		if operand.Type == Operand_Register {
			if operand.Register.RegisterIndex == Register_a {
				return true
			}
		}
	}
	return false
}

func getEAVal(operand InstructionOperand) *EffectiveAddress {
	if operand.Type == Operand_Memory {
		return &operand.EffectiveAddress
	}
	return nil
}

func CalculateInstructionCycles(instruction Instruction, tookJump bool) int {
	opOne := instruction.InstructionOperands[0]
	opTwo := instruction.InstructionOperands[1]
	op1IsRegister := instOperandIsType(opOne, Operand_Register)
	op1IsMemory := instOperandIsType(opOne, Operand_Memory)
	op2IsRegister := instOperandIsType(opTwo, Operand_Register)
	op2IsMemory := instOperandIsType(opTwo, Operand_Memory)
	op2IsImmediate := instOperandIsType(opTwo, Operand_Immediate)

	AccumulatorIsUsed := isAccumulatorUsed(opOne, opTwo)

	cycleTotal := 0

	var eaVal *EffectiveAddress
	if op1IsMemory {
		eaVal = getEAVal(opOne)
	}
	if op2IsMemory {
		eaVal = getEAVal(opTwo)
	}

	switch instruction.Op {
	case Op_mov:
		if op1IsMemory && AccumulatorIsUsed {
			cycleTotal = CalculateCycles(10, 1, eaVal)
		}
		if op1IsRegister && op2IsRegister {
			cycleTotal = CalculateCycles(2, 0, eaVal)
		}
		if op1IsRegister && op2IsMemory {
			cycleTotal = CalculateCycles(8, 1, eaVal)
		}
		if op1IsMemory && op2IsRegister {
			cycleTotal = CalculateCycles(9, 1, eaVal)
		}
		if op1IsMemory && op2IsRegister {
			cycleTotal = CalculateCycles(9, 1, eaVal)
		}
		if op1IsRegister && op2IsImmediate {
			cycleTotal = CalculateCycles(4, 0, eaVal)
		}
		if op1IsMemory && op2IsImmediate {
			cycleTotal = CalculateCycles(10, 1, eaVal)
		}
	case Op_add, Op_sub:
		if op1IsRegister && op2IsRegister {
			cycleTotal = CalculateCycles(3, 0, eaVal)
		}
		if op1IsRegister && op2IsMemory {
			cycleTotal = CalculateCycles(9, 1, eaVal)
		}
		if op1IsMemory && op2IsRegister {
			cycleTotal = CalculateCycles(16, 2, eaVal)
		}
		if op1IsRegister && op2IsImmediate {
			cycleTotal = CalculateCycles(4, 0, eaVal)
		}
		if op1IsMemory && op2IsImmediate {
			cycleTotal = CalculateCycles(17, 2, eaVal)
		}
		if AccumulatorIsUsed && op2IsImmediate {
			cycleTotal = CalculateCycles(4, 0, eaVal)
		}
	// everything below here is janky (inaccurate transfers calc)
	case Op_je, Op_jl, Op_jle, Op_jb, Op_jbe, Op_js, Op_jne, Op_jnl, Op_jg, Op_ja, Op_jns:
		if tookJump {
			cycleTotal = 16
		} else {
			cycleTotal = 4
		}
	case Op_jmp:
		cycleTotal = 15
	case Op_push:
		if op1IsRegister {
			cycleTotal = CalculateCycles(11, 1, eaVal)
		}
		if op1IsMemory {
			cycleTotal = CalculateCycles(16, 2, eaVal)
		}
	case Op_pop:
		if op1IsRegister {
			cycleTotal = CalculateCycles(8, 1, eaVal)
		}
		if op1IsMemory {
			cycleTotal = CalculateCycles(17, 2, eaVal)
		}
	case Op_call:
		if instruction.Size > 2 {
			cycleTotal = 28
		} else {
			cycleTotal = 19
		}
	case Op_ret:
		// intra no pop
		cycleTotal = 8
	}
	return cycleTotal
}
