package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/user/coin-indexer/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the GraphQL API server",
	Long:  `Start the GraphQL server to handle queries for indexed transaction data`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting GraphQL server...")
		
		srv, err := server.NewServer()
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}