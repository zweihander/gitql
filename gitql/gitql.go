package gitql

import (
	"fmt"
	"reflect"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/k0kubun/pp"
	"vitess.io/vitess/go/vt/sqlparser"
)

type Gitql struct {
	repo *git.Repository
}

func New(repo *git.Repository) *Gitql {
	return &Gitql{
		repo: repo,
	}
}

func (gql *Gitql) checkout(branch string) error {
	w, err := gql.repo.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
}

func (gql *Gitql) ExecuteQuery(q string) (*Table, error) {
	stmt, err := sqlparser.Parse(q)
	if err != nil {
		return nil, err
	}

	switch s := stmt.(type) {
	case *sqlparser.Select:
		if len(s.From) != 1 {
			return nil, fmt.Errorf("multiple select not supported")
		}

		tableExpr, ok := s.From[0].(*sqlparser.AliasedTableExpr)
		if !ok {
			return nil, fmt.Errorf("unsupported select from type: %s", reflect.TypeOf(s.From[0]))
		}

		table, ok := tableExpr.Expr.(sqlparser.TableName)
		if !ok {
			return nil, fmt.Errorf("unsupported select table expr: %s", reflect.TypeOf(tableExpr.Expr))
		}

		switch table.Name.String() {
		case "commits":
			return gql.walkCommits(s)
		case "refs", "tags", "branches":
			fallthrough
		default:
			pp.Println(s)
			return nil, fmt.Errorf("unknown table: %s", table.Name.String())
		}
	// case *sqlparser.Use:
	// 	return nil, gql.checkout(s.DBName.String())
	// case *sqlparser.Show:

	default:
		pp.Println(stmt)
		panic("wow")
	}
}
