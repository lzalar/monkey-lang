package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(left, right, node.Operator)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			return newError("identifier not found: %s", node.Value)
		}
		return val
	}

	return nil
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	var returnValue object.Object
	if isTruthy(condition) {
		returnValue = Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		returnValue = Eval(node.Alternative, env)
	} else {
		return NULL
	}
	return returnValue
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

func evalInfixExpression(left object.Object, right object.Object, operator string) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError("type mismatch: %s + %s", left.Type(), right.Type())
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		leftValue := left.(*object.Integer)
		rightValue := right.(*object.Integer)
		return evalIntegerInfixExpression(leftValue, rightValue, operator)
	case operator == token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case operator == token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(left *object.Integer, right *object.Integer, operator string) object.Object {
	switch operator {
	case token.PLUS:
		return &object.Integer{Value: left.Value + right.Value}
	case token.MINUS:
		return &object.Integer{Value: left.Value - right.Value}
	case token.ASTERISK:
		return &object.Integer{Value: left.Value * right.Value}
	case token.SLASH:
		return &object.Integer{Value: left.Value / right.Value}
	case token.EQ:
		return nativeBoolToBooleanObject(left.Value == right.Value)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(left.Value != right.Value)
	case token.LT:
		return nativeBoolToBooleanObject(left.Value < right.Value)
	case token.GT:
		return nativeBoolToBooleanObject(left.Value > right.Value)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalPrefixExpression(operator string, object object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evalBangOperator(object)
	case token.MINUS:
		return evalMinusPrefixOperator(object)
	default:
		return newError("unknown operator: %s%s", operator, object.Type())
	}
}

func evalMinusPrefixOperator(o object.Object) object.Object {
	if o.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: %s%s", token.MINUS, o.Type())
	}

	return &object.Integer{
		Value: -o.(*object.Integer).Value,
	}
}

func evalBangOperator(o object.Object) object.Object {
	switch o {
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

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalBlockStatement(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)

		if result != nil {
			switch result.Type() {
			case object.RETURN_VALUE_OBJ, object.ERROR_OBJ:
				return result
			}
		}
	}

	return result
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
