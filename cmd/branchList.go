package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

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

		// TODO: Add options to control whether info is retrieved from local metadata cache or from the DBHub.io server

		// If there is a local metadata cache for the requested database, use that
		db := args[0]
		md, err := ioutil.ReadFile(filepath.Join(".dio", db, "metadata.json"))
		if err != nil {
			// No local cache, so retrieve the info from the server
			temp, err := retrieveMetadata(db)
			if err != nil {
				return err
			}
			md = []byte(temp)
		}
		list := metaData{}
		err = json.Unmarshal([]byte(md), &list)
		if err != nil {
			return err
		}

		// Sort the list alphabetically
		var sortedKeys []string
		for k := range list.Branches {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Display the list of branches
		fmt.Printf("Branches for %s:\n\n", db)
		for _, i := range sortedKeys {
			fmt.Printf("  * %s - Commit: %s\n", i, list.Branches[i].Commit)
			if list.Branches[i].Description != "" {
				fmt.Printf("\n      %s\n\n", list.Branches[i].Description)
			}
		}
		fmt.Printf("\n    Default branch: %s\n\n", list.DefBranch)
		return nil
	},
}

func init() {
	branchCmd.AddCommand(branchListCmd)
}
