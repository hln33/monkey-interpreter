package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) != 0 {
		t.Errorf("parser has %d errors", len(errors))
		for _, msg := range errors {
			t.Errorf("parser error: %q", msg)
		}
		t.FailNow()
	}
}
func parseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	return program
}
func testNumProgramStatements(t *testing.T, program *ast.Program, expected int) {
	if len(program.Statements) != expected {
		t.Fatalf("program.Statements does not contain %d statement(s). got=%d",
			expected, len(program.Statements))
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)
		testNumProgramStatements(t, program, 1)

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, test.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, test.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, expectedName string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != expectedName {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", expectedName, letStmt.Name.Value)
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return 10;", 10},
		{"return 993322;", 993322},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)
		testNumProgramStatements(t, program, 1)

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ReturnStatement, got=%T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return, got %q",
				returnStmt.TokenLiteral())
		}
		if !testLiteralExpression(t, returnStmt.ReturnValue, test.expectedValue) {
			return
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral() not %s. got=%s",
			"foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5", literal.TokenLiteral())
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input        string
		expectedBool bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)
		testNumProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got=%T", program.Statements[0])
		}

		if !testBooleanLiteral(t, stmt.Expression, test.expectedBool) {
			return
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false", "!", false},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)
		testNumProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		expr, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("expr is not *ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if expr.Operator != test.operator {
			t.Fatalf("expr.Operator is not '%s'. got=%s",
				test.operator, expr.Operator)
		}
		if !testLiteralExpression(t, expr.Right, test.value) {
			return
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)
		testNumProgramStatements(t, program, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, test.leftValue, test.operator, test.rightValue) {
			return
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T",
			stmt.Expression)
	}

	if !testInfixExpression(t, expr.Condition, "x", "<", "y") {
		return
	}

	if len(expr.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n",
			len(expr.Consequence.Statements))
	}

	consequence, ok := expr.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Consequence.Statements[0] is not ast.ExpressionStatement. got=%T",
			expr.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if expr.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", expr.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got %T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.IfExpression, got %T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequnce is not 1 statement, got %d", len(exp.Consequence.Statements))
	}
	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Consequence.Statements[0] is not *ast.ExpressionStatement, got %T", exp.Consequence.Statements[0])
	}
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statement, got %d", len(exp.Alternative.Statements))
	}
	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Alternate.Statements[0] is not *ast.ExpressionStatement, got %T", exp.Alternative.Statements[0])
	}
	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"1 + 2 * 3",
			"(1 + (2 * 3))",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)

		actual := program.String()
		if actual != test.expected {
			t.Errorf("expected=%q, got=%q", test.expected, actual)
		}
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `fn(x, y) { x + y; }`
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	fl, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(fl.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(fl.Parameters))
	}
	testLiteralExpression(t, fl.Parameters[0], "x")
	testLiteralExpression(t, fl.Parameters[1], "y")

	if len(fl.Body.Statements) != 1 {
		t.Fatalf("fl.Body.Statements does not contain 1 statement. got=%d\n",
			len(fl.Body.Statements))
	}
	bodyStmt, ok := fl.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			fl.Body.Statements[0])
	}
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameters(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, test := range tests {
		program := parseProgram(t, test.input)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(test.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(test.expectedParams), len(function.Parameters))
		}

		for i, ident := range test.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	program := parseProgram(t, input)
	testNumProgramStatements(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, expr.Function, "add") {
		return
	}

	if len(expr.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(expr.Arguments))
	}

	testLiteralExpression(t, expr.Arguments[0], 1)
	testInfixExpression(t, expr.Arguments[1], 2, "*", 3)
	testInfixExpression(t, expr.Arguments[2], 4, "+", 5)
}

func testIdentifier(t *testing.T, expr ast.Expression, expectedVal string) bool {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Errorf("expr not *ast.Identifier. got=%T", expr)
		return false
	}

	if ident.TokenLiteral() != expectedVal {
		t.Errorf("ident.TokenLiteral() not %s. got=%s",
			expectedVal, ident.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, expr ast.Expression, expectedVal int64) bool {
	lit, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("expr not *ast.IntegerLiteral. got=%T", expr)
		return false
	}

	if lit.Value != expectedVal {
		t.Errorf("il.Value not %d. got=%d", expectedVal, lit.Value)
	}

	if lit.TokenLiteral() != fmt.Sprintf("%d", expectedVal) {
		t.Errorf("il.TokenLiteral() not %d. got=%s",
			expectedVal, lit.TokenLiteral())
	}

	return true
}

func testBooleanLiteral(t *testing.T, expr ast.Expression, expectedVal bool) bool {
	b, ok := expr.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean, got %T", expr)
		return false
	}

	if b.Value != expectedVal {
		t.Errorf("b.Value not %t, got %t", expectedVal, b.Value)
		return false
	}

	if b.TokenLiteral() != fmt.Sprintf("%t", expectedVal) {
		t.Errorf("b.TokenLiteral not %t, got %s", expectedVal, b.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	expr ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, expr, int64(v))
	case int64:
		return testIntegerLiteral(t, expr, v)
	case string:
		return testIdentifier(t, expr, v)
	case bool:
		return testBooleanLiteral(t, expr, v)
	default:
		t.Errorf("type of expr not handled. got=%T", expr)
		return false
	}
}

func testInfixExpression(
	t *testing.T,
	expr ast.Expression,
	expectedLeft interface{},
	expectedOp string,
	expectedRight interface{},
) bool {
	inExpr, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expr is not *ast.InfixExpression. got=%T(%s)", expr, expr)
	}

	if !testLiteralExpression(t, inExpr.Left, expectedLeft) {
		return false
	}

	if inExpr.Operator != expectedOp {
		t.Errorf("expr.Operator is not '%s'. got=%q", expectedOp, inExpr.Operator)
		return false
	}

	if !testLiteralExpression(t, inExpr.Right, expectedRight) {
		return false
	}

	return true
}
