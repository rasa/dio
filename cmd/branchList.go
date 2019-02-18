package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"

	rq "github.com/parnurzeal/gorequest"
	"github.com/spf13/cobra"
)

// Displays the list of branches for a remote database
var branchListCmd = &cobra.Command{
	Use:   "list [database name]",
	Short: "List the branches for your database on a DBHub.io cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure a database file was given
		if len(args) == 0 {
			return errors.New("No database file specified")
		}
		// TODO: Allow giving multiple database files on the command line.  Hopefully just needs turning this
		// TODO  into a for loop
		if len(args) > 1 {
			return errors.New("Only one database can be worked with at a time (for now)")
		}

		// Retrieve the list of branches
		file := args[0]
		resp, body, errs := rq.New().TLSClientConfig(&TLSConfig).Get(cloud + "/branch/list").
			Query(fmt.Sprintf("username=%s", url.QueryEscape(certUser))).
			Query(fmt.Sprintf("folder=%s", "/")).
			Query(fmt.Sprintf("dbname=%s", url.QueryEscape(file))).
			End()
		if errs != nil {
			log.Print("Errors when retrieving branch list:")
			for _, err := range errs {
				log.Print(err.Error())
			}
			return errors.New("Error when retrieving branch list")
		}
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				return errors.New("Requested database not found")
			}
			return errors.New(fmt.Sprintf("Branch list failed with an error: HTTP status %d - '%v'\n",
				resp.StatusCode, resp.Status))
		}
		list := struct {
			Def     string                 `json:"default_branch"`
			Entries map[string]branchEntry `json:"branches"`
		}{}
		err := json.Unmarshal([]byte(body), &list)
		if err != nil {
			return err
		}

		// Sort the list alphabetically
		var sortedKeys []string
		for k := range list.Entries {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Display the list of branches
		fmt.Printf("Branches for %s:\n\n", file)
		for _, i := range sortedKeys {
			fmt.Printf("  * %s - Commit: %s\n", i, list.Entries[i].Commit)
			if list.Entries[i].Description != "" {
				fmt.Printf("\n      %s\n\n", list.Entries[i].Description)
			}
		}
		fmt.Printf("\n    Default branch: %s\n", list.Def)
		return nil
	},
}

func init() {
	branchCmd.AddCommand(branchListCmd)
}
