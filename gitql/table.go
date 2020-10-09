package gitql

type Table struct {
	Columns []string
	Rows    [][]interface{}
}

func Tables() map[string][]string {
	return map[string][]string{
		"commits": {
			"author_name",
			"author_email",
			"committer_name",
			"committer_email",
			"hash",
			"date",
			"message",
		},
	}
}
