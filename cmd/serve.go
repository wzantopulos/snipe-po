package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wzantopulos/snipe-po/server"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long:  `Start the web UI server for managing purchase orders.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting snipe-po web UI on port %d...\n", servePort)
	return server.Start(servePort)
}
