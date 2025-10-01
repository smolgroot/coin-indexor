package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/user/coin-indexer/internal/indexer"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Start indexing blockchain events",
	Long:  `Start monitoring and indexing events from configured token contracts`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting blockchain event indexer...")
		
		indexerService, err := indexer.NewIndexer()
		if err != nil {
			log.Fatalf("Failed to create indexer: %v", err)
		}
		
		if err := indexerService.Start(); err != nil {
			log.Fatalf("Failed to start indexer: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}