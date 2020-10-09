package gitql

import (
	"github.com/go-git/go-git/v5/plumbing"
	"vitess.io/vitess/go/vt/sqlparser"
)

func (gql *Gitql) walkTags(sel *sqlparser.Select) (*Table, error) {
	tags, err := gql.repo.Tags()
	if err != nil {
		return nil, err
	}

	if err := tags.ForEach(func(tag *plumbing.Reference) error {

		return nil
	}); err != nil && err != errDone {
		return nil, err
	}

	return nil, nil
}
