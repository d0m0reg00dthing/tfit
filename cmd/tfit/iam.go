package main

import (
	"github.com/spf13/cobra"
)

func NewCmdIAM() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "IAM Related",
	}

	cmd.AddCommand(NewCmdIAMPolicy())
	cmd.AddCommand(NewCmdIAMRole())
	cmd.AddCommand(NewCmdIAMUser())
	cmd.AddCommand(NewCmdIAMGroup())

	return cmd
}
