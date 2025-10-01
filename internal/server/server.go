package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	
	"github.com/user/coin-indexer/internal/database"
	"github.com/user/coin-indexer/internal/graphql"
	"github.com/user/coin-indexer/internal/models"
)

type Server struct {
	router *gin.Engine
}

// NewServer creates a new GraphQL server
func NewServer() (*Server, error) {
	// Initialize database
	if err := database.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	
	// Enhanced CORS middleware for GraphQL
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})
	
	// GraphQL endpoint with enhanced configuration
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: &graphql.Resolver{},
	}))
	
	// Handle GraphQL requests
	router.POST("/graphql", gin.WrapH(srv))
	
	// GraphQL playground (enhanced GraphiQL-style interface)
	router.GET("/playground", gin.WrapH(playground.Handler("Coin Indexer - GraphQL Playground", "/graphql")))
	router.GET("/graphiql", gin.WrapH(playground.Handler("Coin Indexer - GraphiQL", "/graphql")))
	
	// REST endpoints
	router.GET("/", rootHandler)
	router.POST("/contracts", addContractHandler)
	router.GET("/health", healthHandler)
	
	return &Server{router: router}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	host := viper.GetString("server.host")
	port := viper.GetString("server.port")
	addr := fmt.Sprintf("%s:%s", host, port)
	
	log.Printf("🚀 GraphQL server starting on %s", addr)
	log.Printf("📊 GraphQL playground available at:")
	log.Printf("   • http://%s/playground (Main GraphQL Playground)", addr)
	log.Printf("   • http://%s/graphiql (Alternative GraphiQL interface)", addr)
	log.Printf("   • http://%s/graphql (GET for playground, POST for queries)", addr)
	log.Printf("💡 Health check: http://%s/health", addr)
	log.Printf("🔗 Add contracts: POST http://%s/contracts", addr)
	
	return s.router.Run(addr)
}

// addContractHandler handles adding new contracts to monitor
func addContractHandler(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Address    string `json:"address" binding:"required"`
		StartBlock uint64 `json:"start_block"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("DEBUG: Received request: %+v", req)
	
	// Create new contract record
	contract := models.Contract{
		Name:       req.Name,
		Address:    req.Address,
		StartBlock: req.StartBlock,
		LastBlock:  0,
		IsActive:   true,
	}
	
	log.Printf("DEBUG: Created contract model: %+v", contract)
	
	// Save to database
	db := database.GetDB()
	log.Printf("DEBUG: Database instance: %v (nil? %t)", db, db == nil)
	
	if err := db.Create(&contract).Error; err != nil {
		log.Printf("DEBUG: Database create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save contract to database"})
		return
	}
	
	log.Printf("DEBUG: Contract saved successfully with ID: %d", contract.ID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Contract added successfully",
		"contract": req,
	})
}

// rootHandler provides API information and links
func rootHandler(c *gin.Context) {
	host := c.Request.Host
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"service": "Coin Indexer API",
		"version": "1.0.0",
		"endpoints": gin.H{
			"graphql_playground": fmt.Sprintf("%s://%s/playground", protocol, host),
			"graphiql":          fmt.Sprintf("%s://%s/graphiql", protocol, host),
			"graphql_endpoint":  fmt.Sprintf("%s://%s/graphql", protocol, host),
			"health_check":      fmt.Sprintf("%s://%s/health", protocol, host),
			"add_contract":      fmt.Sprintf("%s://%s/contracts", protocol, host),
		},
		"sample_queries": []gin.H{
			{
				"name": "Recent Transactions",
				"query": `query {
  transactions(limit: 5) {
    id
    txHash
    fromAddress
    toAddress
    amount
    tokenName
    blockTimestamp
  }
}`,
			},
			{
				"name": "Get Contracts",
				"query": `query {
  contracts {
    id
    name
    address
    isActive
  }
}`,
			},
		},
	})
}

// healthHandler provides a health check endpoint
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "coin-indexer",
	})
}