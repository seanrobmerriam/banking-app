package main

import (
	"banking-app/database"
	"banking-app/handlers"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Application configuration - easily configurable via environment variables
const (
	DefaultPort = "8080" // Default HTTP port if not specified
)

func main() {
	// Initialize database connection
	// Critical first step - application cannot function without database
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			sqlDB.Close()
		}
	}()

	// Initialize HTTP router with middleware
	// Gin provides high-performance routing with minimal overhead
	router := gin.Default()
	
	// CORS middleware for cross-origin requests
	// Essential for web application frontends communicating with backend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint - crucial for monitoring and load balancers
	// Provides basic application status information
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "banking-app",
		})
	})

	// API versioning - important for backward compatibility
	v1 := router.Group("/api/v1")
	{
		// Customer management endpoints - core banking functionality
		customers := v1.Group("/customers")
		{
			customers.GET("", handlers.GetCustomers(db))              // List all customers
			customers.GET(":id", handlers.GetCustomer(db))            // Get customer by ID
			customers.POST("", handlers.CreateCustomer(db))           // Create new customer
			customers.PUT(":id", handlers.UpdateCustomer(db))         // Update customer
			customers.DELETE(":id", handlers.DeleteCustomer(db))      // Delete customer
		}

		// Account management endpoints - core banking functionality
		accounts := v1.Group("/accounts")
		{
			accounts.GET("", handlers.GetAccounts(db))                // List all accounts
			accounts.GET(":id", handlers.GetAccount(db))              // Get account by ID
			accounts.POST("", handlers.CreateAccount(db))             // Create new account
			accounts.PUT(":id", handlers.UpdateAccount(db))           // Update account
			accounts.DELETE(":id", handlers.DeleteAccount(db))        // Delete account
			
			// Account-specific operations
			accounts.GET(":id/balance", handlers.GetAccountBalance(db)) // Get account balance
			accounts.GET(":id/transactions", handlers.GetAccountTransactions(db)) // Get transaction history
		}

		// Transaction processing endpoints - core banking functionality
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", handlers.GetTransactions(db))        // List all transactions
			transactions.POST("", handlers.CreateTransaction(db))     // Process transaction
		}

		// Loan management endpoints - core banking functionality
		loans := v1.Group("/loans")
		{
			loans.GET("", handlers.GetLoans(db))                     // List all loans
			loans.GET(":id", handlers.GetLoan(db))                   // Get loan by ID
			loans.POST("", handlers.CreateLoan(db))                  // Create new loan
			loans.PUT(":id", handlers.UpdateLoan(db))                // Update loan
			loans.DELETE(":id", handlers.DeleteLoan(db))             // Delete loan
		}
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Validate port number - essential for error prevention
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("Invalid port number:", port)
	}

	log.Printf("Core Banking Application starting on port %s", port)
	log.Printf("API Documentation available at: http://localhost:%s/api/v1", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)
	
	// Start HTTP server with graceful shutdown support
	// In production, implement proper signal handling for graceful shutdown
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}