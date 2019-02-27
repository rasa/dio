package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	rq "github.com/parnurzeal/gorequest"
	"github.com/spf13/cobra"
)

var branchUpdateBranch string
var descDel *bool

// Updates the description text for a branch
var branchUpdateCmd = &cobra.Command{
	Use:   "update [database name]",
	Short: "Update the description for a branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure a database file was given
		if len(args) == 0 {
			return errors.New("No database file specified")
		}
		// TODO: Allow giving multiple database files on the command line.  Hopefully just needs turning this
		// TODO  into a for loop
		if len(args) > 1 {
			return errors.New("Only one database can be changed at a time (for now)")
		}

		// Ensure a branch name and description text were given
		if branchUpdateBranch == "" {
			return errors.New("No branch name given")
		}
		if msg == "" && *descDel == false {
			return errors.New("No description text given")
		}

		// Update the branch
		file := args[0]
		req := rq.New().Post(cloud+"/branch_update").
			Set("branch", branchUpdateBranch).
			Set("database", file)
		if msg != "" {
			req.Set("desc", msg)
		} else {
			req.Set("del", "true")
		}
		resp, _, errs := req.End()
		if errs != nil {
			log.Print("Errors when updating branch description:")
			for _, err := range errs {
				log.Print(err.Error())
			}
			return errors.New("Error when updating branch description")
		}
		if resp.StatusCode != http.StatusNoContent {
			if resp.StatusCode == http.StatusNotFound {
				return errors.New("Requested database or branch not found")
			}
			return errors.New(fmt.Sprintf("Description update failed with an error: HTTP status %d - '%v'\n",
				resp.StatusCode, resp.Status))
		}

		// Inform the user
		if *descDel {
			fmt.Println("Description deleted")
		} else {
			fmt.Println("Description updated")
		}
		return nil
	},
}

func init() {
	branchCmd.AddCommand(branchUpdateCmd)
	branchUpdateCmd.Flags().StringVar(&branchUpdateBranch, "branch", "",
		"Name of remote branch to create")
	descDel = branchUpdateCmd.Flags().BoolP("delete", "d", false,
		"Delete the branch description")
	branchUpdateCmd.Flags().StringVar(&msg, "description", "", "Description of the branch")
}
