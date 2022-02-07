package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fatihdumanli/cnote"
	"github.com/fatihdumanli/cnote/internal/survey"
	"github.com/fatihdumanli/cnote/pkg/onenote"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var out io.Writer

var rootCmd = &cobra.Command{
	Long: "Take notes on your Onenote notebooks from terminal",
	Run: func(c *cobra.Command, args []string) {
		os.Exit(startNoteSurvey())
	},
	Use:                   "cnote",
	DisableFlagsInUseLine: true,
}

//The function gets executed once the application starts without any commands/arguments.
func startNoteSurvey() int {
	noteContent, err := survey.AskNoteContent()
	if err != nil {
		panic(err)
	}

	notebookSpinner, _ := pterm.DefaultSpinner.Start("Getting your notebooks...")
	notebooks := cnote.GetNotebooks()
	//TODO: What if it fails, consider use retry.
	notebookSpinner.Success()
	n, err := survey.AskNotebook(notebooks)

	sectionsSpinner, _ := pterm.DefaultSpinner.Start("Getting sections...")
	sections := cnote.GetSections(n)
	//TODO: What if it fails, consider use retry.
	sectionsSpinner.Success()
	section, err := survey.AskSection(n, sections)

	//Saving the note to the section
	_, err = cnote.SaveNotePage(onenote.NotePage{
		Section: section,
		Content: noteContent,
	})
	if err != nil {
		return 1
	}

	a, err := survey.AskAlias(onenote.NotebookName(n.DisplayName), onenote.SectionName(section.Name))
	if a != "" {
		//User answered with an alias
		err := cnote.SaveAlias(a, n, section)
		if err != nil {
			return 2
		}
	}

	return 0
}

func Execute() {
	out = os.Stdout
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
