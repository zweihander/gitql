package gitql

import (
	"fmt"
	"reflect"
	"time"

	"vitess.io/vitess/go/vt/sqlparser"
)

func satisfying(obj dataProvider, expr sqlparser.Expr) (bool, error) {
	switch exp := expr.(type) {
	case *sqlparser.ComparisonExpr:
		ok, err := satisfyingComparison(obj, exp)
		if err != nil {
			return false, fmt.Errorf("comparison: %s", err)
		}
		return ok, nil
	case *sqlparser.FuncExpr:
		fn, ok := builtinFns[exp.Name.Lowered()]
		if !ok {
			return false, fmt.Errorf("unknown function: %s", exp.Name)
		}
		return fn(obj, exp.Exprs, exp.Distinct)
	case *sqlparser.AndExpr:
		l, lerr := satisfying(obj, exp.Left)
		if lerr != nil {
			return false, lerr
		}

		r, rerr := satisfying(obj, exp.Right)
		if rerr != nil {
			return false, rerr
		}

		return (l == true && r == true), nil
	case *sqlparser.OrExpr:
		l, lerr := satisfying(obj, exp.Left)
		if lerr != nil {
			return false, lerr
		}

		r, rerr := satisfying(obj, exp.Right)
		if rerr != nil {
			return false, rerr
		}

		return l || r, nil
	default:
		return false, fmt.Errorf("unsupported expression type: %s", reflect.TypeOf(expr))
	}
}

func satisfyingComparison(obj dataProvider, cmp *sqlparser.ComparisonExpr) (bool, error) {
	col, ok := cmp.Left.(*sqlparser.ColName)
	if !ok {
		return false, fmt.Errorf("unsupported left operand type: %s", reflect.TypeOf(cmp.Left))
	}

	left, err := obj.GetValue(col.Name.Lowered())
	if err != nil {
		return false, err
	}

	sqval, ok := cmp.Right.(*sqlparser.SQLVal)
	if !ok {
		return false, fmt.Errorf("unsupported right operand type: %s", reflect.TypeOf(cmp.Right))
	}

	right, err := decodeSQLVal(sqval)
	if err != nil {
		return false, err
	}

	if lt, rt := reflect.TypeOf(left), reflect.TypeOf(right); lt != rt {
		return false, fmt.Errorf("comparing different types: %s and %s", lt, rt)
	}

	switch cmp.Operator {
	case "=":
		switch l := left.(type) {
		case string:
			return l == right.(string), nil
		case uint64:
			return l == right.(uint64), nil
		case time.Time:
			return l.Equal(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	case "!=":
		switch l := left.(type) {
		case string:
			return l != right.(string), nil
		case uint64:
			return l != right.(uint64), nil
		case time.Time:
			return !l.Equal(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	case ">":
		switch l := left.(type) {
		case string:
			return l > right.(string), nil
		case uint64:
			return l > right.(uint64), nil
		case time.Time:
			return l.After(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	case "<":
		switch l := left.(type) {
		case string:
			return l < right.(string), nil
		case uint64:
			return l < right.(uint64), nil
		case time.Time:
			return l.Before(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	case ">=":
		switch l := left.(type) {
		case string:
			return l >= right.(string), nil
		case uint64:
			return l >= right.(uint64), nil
		case time.Time:
			return l.After(right.(time.Time)) || l.Equal(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	case "<=":
		switch l := left.(type) {
		case string:
			return l <= right.(string), nil
		case uint64:
			return l <= right.(uint64), nil
		case time.Time:
			return l.Before(right.(time.Time)) || l.Equal(right.(time.Time)), nil
		default:
			return false, fmt.Errorf("unsupported comparison type: %s", reflect.TypeOf(left))
		}
	default:
		return false, fmt.Errorf("unsupported comparison operator: %s", cmp.Operator)
	}
}
