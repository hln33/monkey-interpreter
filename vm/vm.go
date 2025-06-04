package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const STACK_SIZE = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack    []object.Object
	stackPtr int // Always points to the next free value. Top of stack is stack[stackPtr-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:    make([]object.Object, STACK_SIZE),
		stackPtr: 0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPtr == 0 {
		return nil
	}
	return vm.stack[vm.stackPtr-1]
}

func (vm *VM) Run() error {
	for ptr := 0; ptr < len(vm.instructions); ptr++ {
		op := code.Opcode(vm.instructions[ptr])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(vm.instructions[ptr+1:])
			ptr += 2

			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()

			leftVal := left.(*object.Integer).Value
			rightVal := right.(*object.Integer).Value
			vm.push(&object.Integer{Value: leftVal + rightVal})
		}
	}

	return nil
}

// pushes an object onto the stack and increments the stack pointer
func (vm *VM) push(o object.Object) error {
	if vm.stackPtr >= STACK_SIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPtr] = o
	vm.stackPtr++

	return nil
}

// pops an object off the stack and decrements the stack pointer
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.stackPtr-1]
	vm.stackPtr--
	return o
}
