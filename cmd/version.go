package cmd

import (
	"docker-visualizer/collector/version"
	"github.com/spf13/cobra"
)

func createVersionCmd(v, commit, branch string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current version info",
		Run: func(cmd *cobra.Command, args []string) {
			version.Info(v, commit, branch	)
		},
	}
}
