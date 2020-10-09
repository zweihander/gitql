package gitql

import (
	"fmt"
	"reflect"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"vitess.io/vitess/go/vt/sqlparser"
)

func (gql *Gitql) walkCommits(sel *sqlparser.Select) (*Table, error) {
	ref, err := gql.repo.Head()
	if err != nil {
		return nil, err
	}

	commits, err := gql.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}

	fields := make([]string, 0, len(sel.SelectExprs))
	for _, e := range sel.SelectExprs {
		switch exp := e.(type) {
		case *sqlparser.AliasedExpr:
			cn, ok := exp.Expr.(*sqlparser.ColName)
			if !ok {
				return nil, fmt.Errorf("unsupported select expr type2: %s", reflect.TypeOf(exp.Expr))
			}

			fields = append(fields, cn.Name.String())
		case *sqlparser.StarExpr:
			fields = Tables()["commits"]
		default:
			return nil, fmt.Errorf("unsupported select expr type: %s", reflect.TypeOf(e))
		}
	}

	var (
		limit, offset uint64
	)
	if sel.Limit != nil {
		lim, hasLimit, err := getLimit(sel.Limit)
		if err != nil {
			return nil, err
		}

		if hasLimit {
			limit = lim
		}

		offs, hasOffset, err := getOffset(sel.Limit)
		if err != nil {
			return nil, err
		}

		if hasOffset {
			offset = offs
		}
	}

	data := make([][]interface{}, 0, limit)
	i := uint64(0)
	if err := commits.ForEach(func(commit *object.Commit) error {
		defer func() { i++ }()
		if i < offset {
			return nil
		}

		shouldBeAppended := false
		if sel.Where != nil {
			satisfies, err := satisfying(commitProvider{commit}, sel.Where.Expr)
			if err != nil {
				return err
			}

			if satisfies {
				shouldBeAppended = true
			}
		} else {
			shouldBeAppended = true
		}

		if shouldBeAppended {
			row, err := extractCommitFields(fields, commit)
			if err != nil {
				return err
			}

			data = append(data, row)
		}

		if limit != 0 && len(data) == int(limit) {
			return errDone
		}
		return nil
	}); err != nil && err != errDone {
		return nil, err
	}

	return &Table{
		Columns: fields,
		Rows:    data,
	}, nil
}

func extractCommitFields(fields []string, commit *object.Commit) ([]interface{}, error) {
	values := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		v, err := commitProvider{commit}.GetValue(field)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}
