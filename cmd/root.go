package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatihdumanli/cnote"
	"github.com/fatihdumanli/cnote/pkg/oauthv2"
	"github.com/fatihdumanli/cnote/pkg/onenote"
	"github.com/fatihdumanli/cnote/storage"
	"github.com/fatihdumanli/cnote/survey"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Long:                  "Take notes on your Onenote notebooks from terminal",
	Run:                   runRoot,
	Use:                   "cnote [command] [args] [flags]",
	DisableFlagsInUseLine: true,
}

type Notebook onenote.Notebook
type Section onenote.Section
type NotebookName = onenote.NotebookName
type SectionName = onenote.SectionName

func runRoot(c *cobra.Command, args []string) {
	var defaultOptions = cnote.GetOptions()

	t, err := getValidAccount()
	if err != nil {
		panic(err)
	}

	_, err = survey.AskNoteContent(defaultOptions)
	if err != nil {
		panic(err)
	}

	notebooks, err := onenote.GetNotebooks(t)
	fmt.Println("Getting your notebooks... This might take a while...")

	if err != nil {
		panic(err)
	}

	n, err := survey.AskNotebook(notebooks)
	sections, err := onenote.GetSections(t, n)
	if err != nil {
		panic(err)
	}

	//iterating over the sections and adding them to the notebook struct
	for _, s := range sections {
		n.Sections = append(n.Sections, s)
	}

	section, err := survey.AskSection(n)

	fmt.Fprintf(defaultOptions.Out, "Your note has saved to the notebook %s and the section %s", n.DisplayName, section.Name)

	a, err := survey.AskAlias(NotebookName(n.DisplayName), SectionName(section.Name))
	if err != nil {
		panic(err)
	}

	if a != "" {
		//TODO: save the alias
		fmt.Println("we need to save the alias.")
	}

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

func getValidAccount() (oauthv2.OAuthToken, error) {
	//TODO: I feel like we shouldn't expose GetOptions() out of the cnote packge.
	var defaultOptions = cnote.GetOptions()
	var oauthParams = getOAuthParams()

	t, st := storage.CheckToken()
	if st == storage.DoesntExist {
		answer, err := survey.AskSetupAccount()
		if !answer || err != nil {
			os.Exit(1)
		}

		token, err := oauthv2.Authorize(oauthParams, defaultOptions.Out)
		if err != nil {
			log.Fatalf("An error occured while trying to authenticate you. %s", err.Error())
		}

		//save the token on local storage
		err = storage.StoreToken(token)
		if err != nil {
			log.Fatalf("An error occured while trying to save the token. %s", err.Error())
		}
	} else if st == storage.Expired {
		//need to refresh the token
		newToken, err := oauthv2.RefreshToken(oauthParams, t.RefreshToken)
		if err != nil {
			panic(err)
		} else {
			//save the token on local storage
			err = storage.StoreToken(newToken)
			if err != nil {
				return t, nil
			}
		}
	}

	return t, nil

}

func getOAuthParams() oauthv2.OAuthParams {
	var msGraphOptions = cnote.GetMicrosoftGraphConfig()
	var oauthParams = oauthv2.OAuthParams{
		OAuthEndpoint:        "https://login.microsoftonline.com/common/oauth2/v2.0",
		RedirectUri:          "http://localhost:5992/oauthv2",
		Scope:                []string{"offline_access", "Notes.ReadWrite.All", "Notes.Create", "Notes.Read", "Notes.ReadWrite"},
		ClientId:             msGraphOptions.ClientId,
		RefreshTokenEndpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
	}
	return oauthParams
}
