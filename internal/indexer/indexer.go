package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	
	"github.com/user/coin-indexer/internal/database"
	"github.com/user/coin-indexer/internal/models"
)

type Indexer struct {
	client     *ethclient.Client
	contracts  []ContractConfig
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

type ContractConfig struct {
	Name       string
	Address    common.Address
	StartBlock uint64
}

// NewIndexer creates a new blockchain indexer
func NewIndexer() (*Indexer, error) {
	providerURL := viper.GetString("blockchain.provider_url")
	if providerURL == "" {
		return nil, fmt.Errorf("blockchain provider URL not configured")
	}
	
	client, err := ethclient.Dial(providerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}
	
	contracts := loadContractsFromConfig()
	
	return &Indexer{
		client:    client,
		contracts: contracts,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start begins indexing for all configured contracts
func (i *Indexer) Start() error {
	log.Println("Starting blockchain indexer...")
	
	// Initialize database connection
	if err := database.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	
	// Start monitoring each contract in its own goroutine
	for _, contract := range i.contracts {
		i.wg.Add(1)
		go i.monitorContract(contract)
	}
	
	// Wait for stop signal
	<-i.stopChan
	
	// Stop all goroutines
	close(i.stopChan)
	i.wg.Wait()
	
	log.Println("Indexer stopped")
	return nil
}

// Stop gracefully stops the indexer
func (i *Indexer) Stop() {
	close(i.stopChan)
}

// monitorContract monitors events for a specific contract
func (i *Indexer) monitorContract(config ContractConfig) {
	defer i.wg.Done()
	
	log.Printf("Starting to monitor contract %s at %s", config.Name, config.Address.Hex())
	
	ticker := time.NewTicker(time.Duration(viper.GetInt("blockchain.poll_interval")) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-i.stopChan:
			log.Printf("Stopping monitor for contract %s", config.Name)
			return
		case <-ticker.C:
			if err := i.processContractEvents(config); err != nil {
				log.Printf("Error processing events for %s: %v", config.Name, err)
			}
		}
	}
}

// processContractEvents processes new events for a contract
func (i *Indexer) processContractEvents(config ContractConfig) error {
	// Get the last processed block for this contract
	lastBlock := i.getLastProcessedBlock(config.Address)
	if lastBlock < config.StartBlock {
		lastBlock = config.StartBlock
	}
	
	// Get current block number
	currentBlock, err := i.client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}
	
	// Skip if no new blocks
	if lastBlock >= currentBlock {
		return nil
	}
	
	// Process blocks in batches
	batchSize := uint64(viper.GetInt("indexing.batch_size"))
	
	for fromBlock := lastBlock + 1; fromBlock <= currentBlock; fromBlock += batchSize {
		toBlock := fromBlock + batchSize - 1
		if toBlock > currentBlock {
			toBlock = currentBlock
		}
		
		if err := i.processBlockRange(config, fromBlock, toBlock); err != nil {
			return fmt.Errorf("failed to process blocks %d-%d: %w", fromBlock, toBlock, err)
		}
		
		// Update last processed block
		i.updateLastProcessedBlock(config.Address, toBlock)
	}
	
	return nil
}

// processBlockRange processes events in a specific block range
func (i *Indexer) processBlockRange(config ContractConfig, fromBlock, toBlock uint64) error {
	// Create filter query for Transfer events
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{config.Address},
		Topics: [][]common.Hash{
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")}, // Transfer event signature
		},
	}
	
	logs, err := i.client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}
	
	for _, vLog := range logs {
		if err := i.processTransferEvent(config, vLog); err != nil {
			log.Printf("Error processing transfer event: %v", err)
			continue
		}
	}
	
	log.Printf("Processed %d events for %s in blocks %d-%d", len(logs), config.Name, fromBlock, toBlock)
	return nil
}

// processTransferEvent processes a single Transfer event
func (i *Indexer) processTransferEvent(config ContractConfig, vLog types.Log) error {
	// Parse Transfer event: Transfer(address indexed from, address indexed to, uint256 value)
	if len(vLog.Topics) < 3 || len(vLog.Data) < 32 {
		return fmt.Errorf("invalid transfer event data")
	}
	
	fromAddress := common.BytesToAddress(vLog.Topics[1].Bytes())
	toAddress := common.BytesToAddress(vLog.Topics[2].Bytes())
	amount := new(big.Int).SetBytes(vLog.Data[:32])
	
	// Get block timestamp
	block, err := i.client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}
	
	// Create transaction record
	tx := &models.Transaction{
		TxHash:          vLog.TxHash.Hex(),
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        uint(vLog.Index),
		ContractAddress: config.Address.Hex(),
		TokenName:       config.Name,
		FromAddress:     fromAddress.Hex(),
		ToAddress:       toAddress.Hex(),
		Amount:          amount.String(),
		BlockTimestamp:  time.Unix(int64(block.Time()), 0),
	}
	
	// Save to database
	db := database.GetDB()
	if err := db.Create(tx).Error; err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}
	
	return nil
}

// getLastProcessedBlock gets the last processed block for a contract
func (i *Indexer) getLastProcessedBlock(address common.Address) uint64 {
	db := database.GetDB()
	var progress models.BlockProgress
	
	if err := db.Where("contract = ?", address.Hex()).First(&progress).Error; err != nil {
		return 0
	}
	
	return progress.LastBlock
}

// updateLastProcessedBlock updates the last processed block for a contract
func (i *Indexer) updateLastProcessedBlock(address common.Address, blockNumber uint64) {
	db := database.GetDB()
	
	progress := models.BlockProgress{
		Contract:  address.Hex(),
		LastBlock: blockNumber,
	}
	
	db.Save(&progress)
}

// loadContractsFromConfig loads contract configurations from config file
func loadContractsFromConfig() []ContractConfig {
	var contracts []ContractConfig
	
	tokens := viper.Get("contracts.tokens")
	if tokens == nil {
		return contracts
	}
	
	tokensList := tokens.([]interface{})
	for _, token := range tokensList {
		tokenMap := token.(map[string]interface{})
		
		name := tokenMap["name"].(string)
		address := common.HexToAddress(tokenMap["address"].(string))
		startBlock := uint64(tokenMap["start_block"].(int))
		
		contracts = append(contracts, ContractConfig{
			Name:       name,
			Address:    address,
			StartBlock: startBlock,
		})
	}
	
	return contracts
}