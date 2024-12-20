/*
Copyright © 2024 Dean
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a rag server",
	Long:  `The serve command starts a gRPC server that provides the rag service.`,
	Run:   RunRAG,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func RunRAG(cmd *cobra.Command, args []string) {
	fmt.Println("serve called")
}
