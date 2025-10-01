# 🪙 Coin Indexer

A high-performance blockchain event indexer that monitors ERC-20 token contracts and provides a powerful GraphQL API for querying transaction data. Perfect for building trading applications, analytics dashboards, and TradingView integrations.

## Features

- 🔗 **Multi-Contract Indexing**: Monitor multiple ERC-20 token contracts simultaneously
- 📊 **Rich Transaction Data**: Index transfers with buyer/seller wallets, amounts, timestamps, and prices
- 🚀 **GraphQL API**: Flexible querying with filtering, pagination, and nested relationships
- ⚙️ **CLI Management**: Simple command-line interface for easy management
- 🔄 **Concurrent Processing**: Goroutine-based monitoring for optimal performance
- 💾 **Database Support**: SQLite for development, PostgreSQL for production
- 📈 **TradingView Ready**: Data format compatible with TradingView Lightweight Charts
- 🔌 **REST Endpoints**: Add new contracts dynamically via API

## Quick Start

### Prerequisites

- Go 1.24+ 
- Ethereum node access (Infura, Alchemy, or local node)

### Installation

```bash
# Clone and setup
git clone https://github.com/user/coin-indexer.git
cd coin-indexer
go mod tidy

# Configure (edit config/config.yaml with your settings)
# Add your Ethereum provider URL and contracts to monitor

# Start indexing blockchain events
go run main.go index

# Start GraphQL server (in another terminal)
go run main.go server
```

Visit `http://localhost:8080/playground` to explore the GraphQL API!

## GraphQL API Examples

### Basic Queries

#### Get Recent Transactions
```graphql
query GetRecentTransactions {
  transactions(limit: 10) {
    id
    txHash
    fromAddress
    toAddress
    amount
    tokenName
    blockTimestamp
    priceUsd
    valueUsd
  }
}
```

#### Get Transactions for Specific Token
```graphql
query GetUSDCTransactions {
  transactions(
    contractAddress: "0xA0b86a33E6417AaF4532CfAd6F41F68481a66CD1"
    limit: 20
  ) {
    id
    txHash
    fromAddress
    toAddress
    amount
    blockNumber
    blockTimestamp
  }
}
```

#### Get Transactions by Address (Sent or Received)
```graphql
query GetAddressActivity {
  addressTransactions(
    address: "0x742d35Cc6634C0532925a3b8D214c5b5c5c4b52"
    limit: 15
  ) {
    id
    txHash
    fromAddress
    toAddress
    amount
    tokenName
    blockTimestamp
  }
}
```

### Advanced Filtering

#### Filter by Block Range
```graphql
query GetTransactionsByBlocks {
  transactions(
    fromBlock: 18500000
    toBlock: 18600000
    limit: 50
  ) {
    id
    txHash
    blockNumber
    fromAddress
    toAddress
    amount
    tokenName
  }
}
```

#### Filter Large Transactions
```graphql
query GetLargeTransfers {
  transactions(
    limit: 25
    tokenName: "USDC"
  ) {
    id
    amount
    valueUsd
    fromAddress
    toAddress
    txHash
    blockTimestamp
  }
}
```

### Contract Information

#### Get All Monitored Contracts
```graphql
query GetContracts {
  contracts {
    id
    name
    address
    startBlock
    lastBlock
    isActive
    createdAt
  }
}
```

#### Get Contract with Recent Transactions
```graphql
query GetContractDetails {
  contract(address: "0xdAC17F958D2ee523a2206206994597C13D831ec7") {
    id
    name
    address
    startBlock
    lastBlock
    transactions {
      id
      amount
      fromAddress
      toAddress
      blockTimestamp
    }
  }
}
```

### Statistics & Analytics

#### Get Transaction Count
```graphql
query GetTransactionStats {
  totalTransactions: transactionCount
  usdcTransactions: transactionCount(
    contractAddress: "0xA0b86a33E6417AaF4532CfAd6F41F68481a66CD1"
  )
}
```

#### Complex Query with Multiple Filters
```graphql
query GetFilteredTransactions {
  transactions(
    tokenName: "USDT"
    fromBlock: 18500000
    limit: 30
    offset: 0
  ) {
    id
    txHash
    blockNumber
    logIndex
    fromAddress
    toAddress
    amount
    priceUsd
    valueUsd
    blockTimestamp
    createdAt
  }
}
```

### Pagination Example

#### Paginated Transaction List
```graphql
query GetPaginatedTransactions($limit: Int!, $offset: Int!) {
  transactions(limit: $limit, offset: $offset) {
    id
    txHash
    fromAddress
    toAddress
    amount
    tokenName
    blockTimestamp
  }
}
```

**Variables:**
```json
{
  "limit": 20,
  "offset": 40
}
```

### Real-time Analytics Queries

#### Recent Activity Dashboard
```graphql
query DashboardData {
  # Recent transactions
  recentTx: transactions(limit: 5) {
    id
    txHash
    amount
    tokenName
    fromAddress
    toAddress
    blockTimestamp
  }
  
  # Contract stats
  contracts {
    name
    address
    lastBlock
    isActive
  }
  
  # Total transaction count
  totalCount: transactionCount
}
```

#### Token-Specific Analytics
```graphql
query TokenAnalytics($tokenContract: String!) {
  # Token transactions
  transactions: transactions(
    contractAddress: $tokenContract
    limit: 100
  ) {
    amount
    valueUsd
    blockTimestamp
  }
  
  # Token info
  tokenInfo: contract(address: $tokenContract) {
    name
    address
    lastBlock
  }
  
  # Transaction count
  txCount: transactionCount(contractAddress: $tokenContract)
}
```

**Variables:**
```json
{
  "tokenContract": "0xA0b86a33E6417AaF4532CfAd6F41F68481a66CD1"
}
```

## 🔌 REST API Examples

### Add New Contract to Monitor
```bash
curl -X POST http://localhost:8080/contracts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WETH",
    "address": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
    "start_block": 18000000
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## ⚙️ Configuration

Edit `config/config.yaml` to configure your setup:

```yaml
# Server settings
server:
  port: "8080"
  host: "localhost"

# Database (SQLite for dev, PostgreSQL for production)
database:
  driver: "sqlite"
  dsn: "coin_indexer.db"

# Blockchain connection
blockchain:
  provider_url: "wss://mainnet.infura.io/ws/v3/YOUR_PROJECT_ID"
  confirmation_blocks: 12
  poll_interval: 15

# Contracts to monitor
contracts:
  tokens:
    - name: "USDC"
      address: "0xA0b86a33E6417AaF4532CfAd6F41F68481a66CD1"
      start_block: 18500000
      abi_file: "abis/erc20.json"
```

## 🏗️ Project Structure

```
coin-indexor/
├── cmd/                    # CLI commands (root, index, server)
├── config/                 # Configuration files
├── internal/
│   ├── database/          # Database connection & migrations
│   ├── indexer/           # Blockchain event monitoring
│   ├── server/            # GraphQL & REST server
│   ├── models/            # Data models
│   └── graphql/           # GraphQL schema & resolvers
├── abis/                  # Contract ABI files
├── .vscode/               # VS Code tasks & debug config
└── main.go               # Application entry point
```

## 🔧 CLI Commands

```bash
# Show available commands
go run main.go --help

# Start blockchain indexer
go run main.go index

# Start GraphQL server
go run main.go server

# Get command-specific help
go run main.go server --help
```

## 🗄️ Database Schema

### Tables
- **transactions** - Individual ERC-20 transfers with metadata
- **contracts** - Monitored token contracts configuration
- **block_progress** - Indexing progress tracking per contract

### Key Fields
- `tx_hash` - Ethereum transaction hash
- `from_address` / `to_address` - Transfer participants
- `amount` - Token amount (as string for big numbers)
- `price_usd` / `value_usd` - Optional USD pricing data
- `block_timestamp` - When transaction occurred
- `contract_address` - Token contract address

## Development

### Adding New Contracts

**Via Config File:**
```yaml
contracts:
  tokens:
    - name: "WETH"
      address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
      start_block: 18000000
      abi_file: "abis/erc20.json"
```

**Via REST API:**
```bash
curl -X POST http://localhost:8080/contracts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WETH",
    "address": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
    "start_block": 18000000
  }'
```

### VS Code Tasks

Available tasks in VS Code:
- **Build Coin Indexer** - Compile the project
- **Run GraphQL Server** - Start API server (background)
- **Start Indexer** - Begin blockchain monitoring (background)

### Debugging

Use VS Code's built-in debugger with the provided launch configurations:
- **Launch GraphQL Server** - Debug API server
- **Launch Indexer** - Debug blockchain indexing
- **Debug Tests** - Run and debug test suite

## Monitoring & Logs

The indexer provides detailed logging for monitoring:

```bash
# View real-time logs
go run main.go index

# Example log output:
# INFO: Starting blockchain indexer...
# INFO: Starting to monitor contract USDC at 0xA0b8...
# INFO: Processed 15 events for USDC in blocks 18500100-18500200
```

##Production Deployment

### PostgreSQL Setup
```yaml
database:
  driver: "postgres"
  dsn: "host=localhost user=indexer password=secret dbname=coin_indexer port=5432 sslmode=disable"
```

### Environment Variables
```bash
export CONFIG_FILE=/path/to/production/config.yaml
go run main.go server --config $CONFIG_FILE
```

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o coin-indexer .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/coin-indexer .
COPY --from=builder /app/config ./config
COPY --from=builder /app/abis ./abis
CMD ["./coin-indexer", "server"]
```

## Performance Tips

- **Batch Size**: Adjust `indexing.batch_size` based on node performance
- **Poll Interval**: Increase `blockchain.poll_interval` for less frequent updates
- **Database**: Use PostgreSQL with proper indexing for production
- **Confirmation Blocks**: Set `confirmation_blocks` based on finality requirements

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Happy indexing! 🚀**

For support or questions, please open an issue on GitHub.