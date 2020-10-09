package gitql

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type dataProvider interface {
	GetValue(field string) (interface{}, error)
}

type commitProvider struct {
	c *object.Commit
}

func (cp commitProvider) GetValue(field string) (interface{}, error) {
	switch field {
	case "author_name":
		return cp.c.Author.Name, nil
	case "author_email":
		return cp.c.Author.Email, nil
	case "committer_name":
		return cp.c.Committer.Name, nil
	case "committer_email":
		return cp.c.Committer.Email, nil
	case "hash":
		return cp.c.Hash.String(), nil
	case "date":
		return cp.c.Author.When, nil // Commiter.When???
	case "message":
		return cp.c.Message, nil
	default:
		return nil, fmt.Errorf("field '%s' does not exist in the commit structure", field)
	}
}
