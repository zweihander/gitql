package gitql

import (
	"fmt"
	"strconv"
	"time"

	"vitess.io/vitess/go/vt/sqlparser"
)

func getLimit(constr *sqlparser.Limit) (uint64, bool, error) {
	if constr.Rowcount == nil {
		return 0, false, nil
	}

	sqlval, ok := constr.Rowcount.(*sqlparser.SQLVal)
	if !ok {
		return 0, false, fmt.Errorf("badbad")
	}

	limit, err := decodeSQLVal(sqlval)
	lim, ok := limit.(uint64)
	return lim, ok, err
}

func getOffset(constr *sqlparser.Limit) (uint64, bool, error) {
	if constr.Offset == nil {
		return 0, false, nil
	}

	sqlval, ok := constr.Offset.(*sqlparser.SQLVal)
	if !ok {
		return 0, false, fmt.Errorf("badbad")
	}

	offset, err := decodeSQLVal(sqlval)
	off, ok := offset.(uint64)
	return off, ok, err
}

const (
	timeYMD    = "2006-01-02"
	timeYMDHIS = "2006-01-02 15:04:05"
)

func decodeSQLVal(val *sqlparser.SQLVal) (interface{}, error) {
	switch val.Type {
	case sqlparser.StrVal:
		s := string(val.Val)
		if t, err := time.Parse(timeYMD, s); err == nil {
			return t, nil
		} else if t, err := time.Parse(timeYMDHIS, s); err == nil {
			return t, nil
		}

		return s, nil
	case sqlparser.IntVal:
		return strconv.ParseUint(string(val.Val), 10, 64)
	case sqlparser.FloatVal:
		return strconv.ParseFloat(string(val.Val), 64)
	default:
		return nil, fmt.Errorf("unsupported value type: %s", val.Val)
	}
}
