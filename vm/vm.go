package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const STACK_SIZE = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.NULL{}

func nativeBoolToBoolObj(input bool) *object.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.NULL:
		return false
	default:
		return true
	}
}

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

// FOR TESTS ONLY
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.stackPtr]
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPtr == 0 {
		return nil
	}
	return vm.stack[vm.stackPtr-1]
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

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpJump:
			jumpPos := int(code.ReadUint16(vm.instructions[ptr+1:]))
			ptr = jumpPos - 1

		case code.OpJumpNotTruthy:
			jumpPos := int(code.ReadUint16(vm.instructions[ptr+1:]))
			ptr += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				ptr = jumpPos - 1
			}

		case code.OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	val := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -val})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}
func (vm *VM) executeBinaryIntegerOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var res int64

	switch op {
	case code.OpAdd:
		res = leftVal + rightVal
	case code.OpSub:
		res = leftVal - rightVal
	case code.OpMul:
		res = leftVal * rightVal
	case code.OpDiv:
		res = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: res})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBoolObj(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBoolObj(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}
func (vm *VM) executeIntegerComparison(
	op code.Opcode,
	left, right object.Object,
) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBoolObj(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBoolObj(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBoolObj(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}
