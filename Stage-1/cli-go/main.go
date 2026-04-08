package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "cogito",
		Short: "Cogito is a persistent memory layer for Ai",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to Cogito. Use --help for commands")
		},
	}

	var initCmd = &cobra.Command{
		Use: "init",
		Short: "Initialize a new Cogito Memo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Cogito initialized!")
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error initing the library: ", err.Error())
		os.Exit(1)
	}
}
