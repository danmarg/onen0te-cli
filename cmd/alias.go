package main

import (
	"fmt"
	"os"
	"sort"

	errors "github.com/pkg/errors"

	"github.com/fatihdumanli/cnote"
	"github.com/fatihdumanli/cnote/internal/style"
	"github.com/fatihdumanli/cnote/internal/survey"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var cmdAlias = &cobra.Command{
	Use:   "alias <command>",
	Short: "add/list and remove alias",
	Long:  "aliases are used for quick access to onenote sections. you can quickly add a new note to any onenote section by specifiying an alias along with your command input.",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "display alias list",
	RunE: func(c *cobra.Command, args []string) error {
		var code, err = displayAliasList()
		os.Exit(code)
		return err
	},
}

var newAliasCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new alias.",
	RunE: func(c *cobra.Command, args []string) error {
		var code, err = newAlias()
		os.Exit(code)
		return err
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"delete"},
	Short:   "remove an alias",
	Args:    cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		os.Exit(removeAlias(c, args))
	},
	DisableFlagsInUseLine: true,
}

func newAlias() (int, error) {
	var notebooks, err = cnote.GetNotebooks()
	if err != nil {
		return 1, errors.Wrap(err, "getNotebooks operation has failed")
	}

	n, err := survey.AskNotebook(notebooks)
	if err != nil {
		return 2, errors.Wrap(err, "askNotebook operation has failed")
	}

	sections, err := cnote.GetSections(n)
	if err != nil {
		return 3, errors.Wrap(err, "getSections operation has failed")
	}

	s, err := survey.AskSection(n, sections)
	if err != nil {
		return 4, errors.Wrap(err, "askSection operation has failed")
	}
	aliasList, err := cnote.GetAliases()
	if err != nil {
		return 1, errors.Wrap(err, "getAliases operation has failed")
	}

	//Check if there's already an alias for the section.
	for _, a := range *aliasList {
		if a.Section.ID == s.ID {
			var warningMsg = fmt.Sprintf("There's already an alias for the section %s. Run cnote alias list to see the whole list.", s.Name)
			fmt.Println(style.Warning(warningMsg))
			return 7, fmt.Errorf("another alias for the section already exists")
		}
	}

	answer, err := survey.AskAlias(s, aliasList)
	if err != nil {
		return 5, errors.Wrap(err, "askAlias operation has failed")
	}
	if answer == "" {
		return 0, nil
	}

	err = cnote.SaveAlias(answer, n, s)
	if err == nil {
		return 0, nil
	}
	return 6, errors.Wrap(err, "saveAlias operation has failed")
}

func displayAliasList() (int, error) {
	var aliasList, err = cnote.GetAliases()
	if err != nil {
		return 1, errors.Wrap(err, "getAliases operation has failed")
	}

	sort.Slice(*aliasList, func(i, j int) bool {
		return (*aliasList)[i].Short < (*aliasList)[j].Short
	})

	if aliasList == nil {
		fmt.Println(style.Error("Your alias data couldn't be loaded."))
		return 1, errors.Wrap(err, "getAliases operation has failed")
	}

	if len(*aliasList) == 0 {
		fmt.Println(style.Error("You haven't added any alias yet."))
		return 2, nil
	}

	var tableData [][]string
	tableData = append(tableData, []string{"Alias", "Section", "Notebook"})

	for _, a := range *aliasList {
		tableData = append(tableData, []string{a.Short, a.Section.Name, a.Notebook.DisplayName})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	return 0, nil
}

func removeAlias(c *cobra.Command, args []string) int {
	if len(args) != 1 {
		c.Usage()
		return 1
	}

	err := cnote.RemoveAlias(args[0])
	if err != nil {
		return 2
	}

	return 0
}

func init() {
	cmdAlias.AddCommand(newAliasCmd)
	cmdAlias.AddCommand(listCmd)
	cmdAlias.AddCommand(removeCmd)
	rootCmd.AddCommand(cmdAlias)
}
