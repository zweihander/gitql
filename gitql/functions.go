package gitql

import (
	"fmt"
	"reflect"
	"strings"

	"vitess.io/vitess/go/vt/sqlparser"
)

type sqlFunction func(obj dataProvider, exprs sqlparser.SelectExprs, distinct bool) (bool, error)

var builtinFns = map[string]sqlFunction{
	"contains": containsSQLFn,
}

// hacky func
func containsSQLFn(obj dataProvider, exprs sqlparser.SelectExprs, distinct bool) (bool, error) {
	if len(exprs) != 2 {
		return false, fmt.Errorf("bad params count")
	}

	field := exprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.ColName).Name.Lowered()
	left, err := obj.GetValue(field)
	if err != nil {
		return false, err
	}

	right, err := decodeSQLVal(exprs[1].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal))
	if err != nil {
		return false, err
	}

	if lt, rt := reflect.TypeOf(left), reflect.TypeOf(right); lt != rt {
		return false, fmt.Errorf("comparing different types: %s and %s", lt, rt)
	}

	switch l := left.(type) {
	case string:
		return strings.Contains(l, right.(string)), nil
	default:
		return false, fmt.Errorf("unsupported operand type: %s", reflect.TypeOf(left))
	}
}
