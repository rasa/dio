package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// branchDefaultGetCmd represents the branchDefaultGet command
var branchDefaultGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the default branch name for a database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("branch default get called")
	},
}

func init() {
	branchDefaultCmd.AddCommand(branchDefaultGetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// branchDefaultGetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// branchDefaultGetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}