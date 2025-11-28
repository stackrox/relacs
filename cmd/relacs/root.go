package main

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "relacs",
	Short: "relacs automates the ACS release process",
	Long:  `relacs is a collection of utility commands to interact with GitHub, Konflux & co.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")
}
