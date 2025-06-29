package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

func (i Instructions) String() string {
	var out bytes.Buffer

	offset := 0
	for offset < len(i) {
		def, err := Lookup(i[offset])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, i[offset+1:])
		fmt.Fprintf(&out, "%04d %s\n", offset, i.fmtInstruction(def, operands))
		offset += 1 + read
	}

	return out.String()
}

func (i Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn // Return with no value. I.e., Null.
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
)

type Definition struct {
	Name          string
	OperandWidths []int // number of bytes each operands takes up
}

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}}, // operand: index of constant
	OpAdd:           {"OpAdd", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // operand: position to jump to
	OpJump:          {"OpJump", []int{2}},          // operand: position to jump to
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}}, // operand: index of global
	OpSetGlobal:     {"OpSetGlobal", []int{2}}, // operand: index of global
	OpArray:         {"OpArray", []int{2}},     // operand: number of elements in array
	OpHash:          {"OpHash", []int{2}},      // operand: number of key AND values on the stack
	OpIndex:         {"OpIndex", []int{}},
	OpCall:          {"OpCall", []int{1}}, // operand: number of arguments
	OpReturnValue:   {"OpReturnValue", []int{}},
	OpReturn:        {"OpReturn", []int{}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
	OpGetBuiltin:    {"OpGetBuiltin", []int{1}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// Create a bytecode instruction with an Opcode and optional number of operands
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// Returns decoded operands from a bytecode instruction and how many bytes it read
func ReadOperands(def *Definition, ins Instructions) (operands []int, bytesRead int) {
	operands = make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	bytesRead = offset
	return operands, bytesRead
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
