package cmd

import "github.com/spf13/cobra"

type CollectorRootCmd struct {
	cobra.Command
	RunCmd     *cobra.Command
	VersionCmd *cobra.Command
}

func CreateRootCmd(name, version, commit, branch string) *CollectorRootCmd {
	rootCmd := &CollectorRootCmd{}
	rootCmd.Use = name
	rootCmd.RunCmd = createRunCmd("collector")
	rootCmd.VersionCmd = createVersionCmd(version, commit, branch)

	rootCmd.AddCommand(rootCmd.RunCmd)
	rootCmd.AddCommand(rootCmd.VersionCmd)

	return rootCmd
}
