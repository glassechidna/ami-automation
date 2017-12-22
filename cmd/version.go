package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/glassechidna/ami-automation/shared"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Output ami-automation version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nBuild Date: %s\n", shared.ApplicationVersion, shared.ApplicationBuildDate)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
