package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const STACK_SIZE = 2048
const GLOBALS_SIZE = 65536
const MAX_FRAMES = 1024

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
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next free value. Top of stack is stack[stackPtr-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MAX_FRAMES)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, STACK_SIZE),
		sp:    0,

		globals: make([]object.Object, GLOBALS_SIZE),

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
}

// FOR TESTS ONLY
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// pushes an object onto the stack and increments the stack pointer
func (vm *VM) push(o object.Object) error {
	if vm.sp >= STACK_SIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// pops an object off the stack and decrements the stack pointer
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

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
			jumpPos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = jumpPos - 1

		case code.OpJumpNotTruthy:
			jumpPos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = jumpPos - 1
			}

		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIdx] = vm.pop()

		case code.OpGetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIdx])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			arr := vm.buildArray(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems

			err := vm.push(arr)
			if err != nil {
				return err
			}

		case code.OpHash:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElems, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElems

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case code.OpIndex:
			idx := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, idx)
			if err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.callFunction(int(numArgs))
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnVal := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnVal)
			if err != nil {
				return err
			}

		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIdx)] = vm.pop()

		case code.OpGetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIdx)])
			if err != nil {
				return err
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

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)

	}
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

func (vm *VM) executeBinaryStringOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return vm.push(&object.String{Value: leftVal + rightVal})
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

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrObj := array.(*object.Array)
	i := index.(*object.Integer).Value

	max := int64(len(arrObj.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrObj.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObj := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildArray(startIdx, endIdx int) object.Object {
	elems := make([]object.Object, endIdx-startIdx)

	for i := startIdx; i < endIdx; i++ {
		elems[i-startIdx] = vm.stack[i]
	}

	return &object.Array{Elements: elems}
}

func (vm *VM) buildHash(startIdx, endIdx int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIdx; i < endIdx; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) callFunction(numArgs int) error {
	fn, ok := vm.stack[vm.sp-1-numArgs].(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("calling non-function")
	}

	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + fn.NumLocals

	return nil
}
