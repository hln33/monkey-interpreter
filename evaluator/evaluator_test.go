package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	res, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if res.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			res.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	res, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if res.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			res.Value, expected)
		return false
	}
	return true
}

func TestIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testIntegerObject(t, eval, test.expected)
	}
}

func TestStringExpression(t *testing.T) {
	input := `"Hello World!"`

	eval := testEval(input)
	str, ok := eval.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", eval, eval)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	eval := testEval(input)
	str, ok := eval.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", eval, eval)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testBooleanObject(t, eval, test.expected)
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	eval := testEval(input)
	res, ok := eval.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", eval, eval)
	}

	if len(res.Elements) != 3 {
		t.Fatalf("array does not have 3 elements. got=%d",
			len(res.Elements))
	}

	testIntegerObject(t, res.Elements[0], 1)
	testIntegerObject(t, res.Elements[1], 4)
	testIntegerObject(t, res.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		integer, ok := test.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!!!5", false},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testBooleanObject(t, eval, test.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		integer, ok := test.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
if (10 > 1) {
	if (10 > 1) {
		return 10;
	}
	return 1;
}
`,
			10,
		},
		{
			`
if (10 > 1) {
	if (1 > 10) {
		return 10;
	}
	return 1;
}
`,
			1,
		},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testIntegerObject(t, eval, test.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testIntegerObject(t, eval, test.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	expectedBody := "(x + 2)"

	eval := testEval(input)
	fn, ok := eval.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", eval, eval)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, test := range tests {
		eval := testEval(test.input)
		testIntegerObject(t, eval, test.expected)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// len(str)
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		// len(arr)
		{"len([])", 0},
		{"len([1])", 1},
		{"len([1, 1 + 2 * 3, true])", 3},
		// first(arr)
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`first(1)`, "argument to `first` must be ARRAY, got INTEGER"},
		{`first(1, 2)`, "wrong number of arguments. got=2, want=1"},
		// last(arr)
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`last(1)`, "argument to `last` must be ARRAY, got INTEGER"},
		{`last(1, 2)`, "wrong number of arguments. got=2, want=1"},
		// rest(arr)
		{`rest([1, 2, 3])`, []int64{2, 3}},
		{`rest([])`, nil},
		{`rest(1)`, "argument to `rest` must be ARRAY, got INTEGER"},
		{`rest(1, 2)`, "wrong number of arguments. got=2, want=1"},
		// push(arr, elem)
		{`push([], 1)`, []int64{1}},
		{`push(1, 2)`, "argument to `push` must be ARRAY, got INTEGER"},
		{`push(1)`, "wrong number of arguments. got=1, want=2"},
	}

	for _, test := range tests {
		eval := testEval(test.input)

		switch expected := test.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			errObj, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					eval, eval)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		case []int64:
			arrObj, ok := eval.(*object.Array)
			if !ok {
				t.Errorf("object is not *object.Array. got=%#v", eval)
				continue
			}
			if len(arrObj.Elements) != len(expected) {
				t.Errorf("wrong number of elements. want=%d, got=%d",
					len(arrObj.Elements), len(expected))
			}
			for i, elem := range arrObj.Elements {
				testIntegerObject(t, elem, expected[i])
			}
		case nil:
			testNullObject(t, eval)
		default:
			t.Errorf("unsupported evaluated value: %#v, want=%#v", eval, test.expected)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1) {
	if (10 > 1) {
		return true + false;
	}
	return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
	}

	for _, test := range tests {
		eval := testEval(test.input)

		errObj, ok := eval.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)",
				eval, eval)
			continue
		}

		if errObj.Message != test.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				test.expectedMessage, errObj.Message)
		}
	}
}
