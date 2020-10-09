package gitql

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"
)

var errDone = fmt.Errorf("done")

func (gql *Gitql) walk(sel *sqlparser.Select) (*Table, error) {
	//sel.SelectExprs
	return nil, nil
}
