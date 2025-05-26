package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.NULL{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func nativeBoolToBoolObj(input bool) *object.Boolean {
	if input {
		return TRUE
	} else {
		return FALSE
	}
}

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BlockStatement:
		return evalBlockStatement(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: val}
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBoolObj(node.Value)
	default:
		return nil
	}
}

func evalProgram(stmts []ast.Statement) object.Object {
	var res object.Object

	for _, stmt := range stmts {
		res = Eval(stmt)

		if rv, ok := res.(*object.ReturnValue); ok {
			return rv.Value
		}
	}

	return res
}

func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var res object.Object

	for _, stmt := range block.Statements {
		res = Eval(stmt)

		if res != nil && res.Type() == object.RETURN_VALUE_OBJ {
			return res
		}
	}

	return res
}

func evalIfExpression(ie *ast.IfExpression) object.Object {
	cond := Eval(ie.Condition)
	if isTruthy(cond) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return NULL
	}
}
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NULL
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(
	op string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right)
	case op == "==":
		return nativeBoolToBoolObj(left == right)
	case op == "!=":
		return nativeBoolToBoolObj(left != right)
	default:
		return NULL
	}
}
func evalIntegerInfixExpression(
	op string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBoolObj(leftVal < rightVal)
	case ">":
		return nativeBoolToBoolObj(leftVal > rightVal)
	case "==":
		return nativeBoolToBoolObj(leftVal == rightVal)
	case "!=":
		return nativeBoolToBoolObj(leftVal != rightVal)
	default:
		return NULL
	}
}
