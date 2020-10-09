package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"
	"github.com/cloudson/gitql/gitql"
	"github.com/go-git/go-git/v5"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

// Version of Gitql
const Version = "Gitql 2.1.0"

func main() {
	app := &cli.App{
		Name:        "gitql",
		Usage:       "A git query language",
		Version:     Version,
		HideVersion: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Enter to interactive mode",
			},
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Value:   ".",
				Usage:   `The (optional) path to run gitql`,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "table",
				Usage:   "The output type format {table|json}",
			},
			// for backward compatibility
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Hidden:  true,
			},
			&cli.StringFlag{
				Name:   "type",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:   "show-tables",
				Hidden: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "show-tables",
				Aliases: []string{"s"},
				Usage:   "Show all tables",
				Action:  showTablesCmd,
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "The version of gitql",
				Action: func(c *cli.Context) error {
					fmt.Println(Version)
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			path, format, interactive := c.String("path"), c.String("format"), c.Bool("interactive")

			// for backward compatibility
			if c.Bool("version") {
				fmt.Println(Version)
				return nil
			}

			if c.Bool("show-tables") {
				return showTablesCmd(c)
			}

			if typ := c.String("type"); typ != "" {
				format = typ
			}
			// ============================

			if c.NArg() == 0 && !interactive {
				return cli.ShowAppHelp(c)
			}

			repo, err := git.PlainOpen(path)
			if err != nil {
				return err
			}

			gql := gitql.New(repo)

			if interactive {
				return runPrompt(gql, format == "json")
			}

			res, err := gql.ExecuteQuery(c.Args().First())
			if err != nil {
				return err
			}

			if format == "json" {
				printJSON(res)
				return nil
			}

			printTable(res)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func showTablesCmd(c *cli.Context) error {
	fmt.Print("Tables: \n\n")

	for tableName, fields := range gitql.Tables() {
		fmt.Printf("%s\n\t", tableName)
		for i, field := range fields {
			comma := "."
			if i+1 < len(fields) {
				comma = ", "
			}
			fmt.Printf("%s%s", field, comma)
		}
		fmt.Println()
	}
	return nil
}

func runPrompt(gql *gitql.Gitql, jsonify bool) error {
	term, err := readline.NewEx(&readline.Config{
		Prompt:       "gitql> ",
		AutoComplete: readline.SegmentFunc(suggestQuery),
	})
	if err != nil {
		return err
	}
	defer term.Close()

	for {
		query, err := term.Readline()
		if err != nil {
			if err == io.EOF {
				break // Ctrl^D
			}
			return err
		}

		if query == "" {
			continue
		}

		if query == "exit" || query == "quit" {
			break
		}

		res, err := gql.ExecuteQuery(query)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			continue
		}

		if jsonify {
			printJSON(res)
			continue
		}

		printTable(res)
	}

	return nil
}

func printTable(result *gitql.Table) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(result.Columns)
	table.SetRowLine(true)
	for _, row := range result.Rows {
		rowData := make([]string, len(row))
		for i, rowval := range row {
			rowData[i] = fmt.Sprintf("%v", rowval)
		}
		table.Append(rowData)
	}
	table.Render()
}

func printJSON(table *gitql.Table) {
	json.NewEncoder(os.Stdout).Encode(table)
}
