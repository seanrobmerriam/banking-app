package handlers

import (
	"banking-app/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Utility function to generate unique account numbers
// Essential for banking systems - each account must have a unique identifier
func generateAccountNumber() string {
	return "ACC" + time.Now().Format("20060102150405") + strconv.Itoa(int(time.Now().UnixNano()%1000))
}

// Utility function to generate unique transaction IDs
// Critical for audit trails and transaction tracking
func generateTransactionID() string {
	return "TXN" + time.Now().Format("20060102150405") + strconv.Itoa(int(time.Now().UnixNano()%1000))
}

// Utility function to generate unique loan numbers
// Important for loan tracking and regulatory compliance
func generateLoanNumber() string {
	return "LOAN" + time.Now().Format("20060102150405") + strconv.Itoa(int(time.Now().UnixNano()%1000))
}

// ==================== CUSTOMER HANDLERS ====================

// GetCustomers retrieves all customers with pagination support
// Important for customer management and regulatory reporting
func GetCustomers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		// Query customers with pagination
		var customers []models.Customer
		var total int64

		db.Model(&models.Customer{}).Count(&total)
		err := db.Preload("Accounts").Preload("Loans").Offset(offset).Limit(limit).Find(&customers).Error
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve customers"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"customers": customers,
			"total":     total,
			"page":      page,
			"limit":     limit,
		})
	}
}

// GetCustomer retrieves a single customer by ID with related data
// Essential for customer service and account access
func GetCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		var customer models.Customer
		err = db.Preload("Accounts").Preload("Loans").First(&customer, uint(id)).Error
		
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, customer)
	}
}

// CreateCustomer creates a new customer record
// Core banking function - first step in customer onboarding
func CreateCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var customer models.Customer
		
		// Validate and bind JSON request
		if err := c.ShouldBindJSON(&customer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
			return
		}

		// Business validation - email uniqueness is handled by database constraint
		if customer.FirstName == "" || customer.LastName == "" || customer.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "First name, last name, and email are required"})
			return
		}

		// Set default values
		customer.Status = "active"
		
		// Create customer record
		if err := db.Create(&customer).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Customer created successfully",
			"customer": customer,
		})
	}
}

// UpdateCustomer updates existing customer information
// Important for customer data maintenance and regulatory compliance
func UpdateCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		var customer models.Customer
		
		// Verify customer exists
		if err := db.First(&customer, uint(id)).Error; err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}

		var updateData models.Customer
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Update customer information
		if err := db.Model(&customer).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Customer updated successfully",
			"customer": customer,
		})
	}
}

// DeleteCustomer performs soft delete of customer record
// Important for data retention policies and audit trails
func DeleteCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
			return
		}

		// Check for active accounts before deletion
		var activeAccounts int64
		db.Model(&models.Account{}).Where("customer_id = ? AND status = 'active'", uint(id)).Count(&activeAccounts)
		
		if activeAccounts > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete customer with active accounts"})
			return
		}

		err = db.Delete(&models.Customer{}, uint(id)).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
	}
}

// ==================== ACCOUNT HANDLERS ====================

// GetAccounts retrieves all accounts with customer information
// Essential for account management and reporting
func GetAccounts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		var accounts []models.Account
		var total int64

		db.Model(&models.Account{}).Count(&total)
		err := db.Preload("Customer").Offset(offset).Limit(limit).Find(&accounts).Error
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accounts": accounts,
			"total":    total,
			"page":     page,
			"limit":    limit,
		})
	}
}

// GetAccount retrieves a single account with transaction history
func GetAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		var account models.Account
		err = db.Preload("Customer").Preload("Transactions").First(&account, uint(id)).Error
		
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}

		c.JSON(http.StatusOK, account)
	}
}

// CreateAccount creates a new bank account for an existing customer
// Core banking function - account opening process
func CreateAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var account models.Account
		
		if err := c.ShouldBindJSON(&account); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate customer exists
		var customer models.Customer
		if err := db.First(&customer, account.CustomerID).Error; err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}

		// Set default values and generate account number
		account.AccountNumber = generateAccountNumber()
		account.Balance = 0.0
		account.Currency = "USD"
		account.Status = "active"

		if err := db.Create(&account).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Account created successfully",
			"account": account,
		})
	}
}

// GetAccountBalance retrieves current balance for an account
// Critical for real-time balance inquiries
func GetAccountBalance(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		var account models.Account
		err = db.Select("id, account_number, balance, currency, status").First(&account, uint(id)).Error
		
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"account_id":    account.ID,
			"account_number": account.AccountNumber,
			"balance":       account.Balance,
			"currency":      account.Currency,
			"status":        account.Status,
		})
	}
}

// GetAccountTransactions retrieves transaction history for an account
// Important for account statements and audit trails
func GetAccountTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
			return
		}

		var transactions []models.Transaction
		err = db.Where("account_id = ?", uint(id)).Order("created_at DESC").Find(&transactions).Error
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"account_id":   uint(id),
			"transactions": transactions,
		})
	}
}

// ==================== TRANSACTION HANDLERS ====================

// CreateTransaction processes financial transactions (deposits, withdrawals)
// Core banking function - money movement processing
func CreateTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transaction models.Transaction
		
		if err := c.ShouldBindJSON(&transaction); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate transaction type
		validTypes := []string{"deposit", "withdrawal", "transfer", "payment"}
		if !contains(validTypes, transaction.TransactionType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction type"})
			return
		}

		// Validate amount is positive
		if transaction.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction amount must be positive"})
			return
		}

		// Get account and perform transaction in database transaction for atomicity
		err := db.Transaction(func(tx *gorm.DB) error {
			var account models.Account
			
			if err := tx.First(&account, transaction.AccountID).Error; err != nil {
				return err
			}

			// Check account status
			if account.Status != "active" {
				return gorm.ErrInvalidData
			}

			// Store balance before transaction
			transaction.BalanceBefore = account.Balance

			// Process transaction based on type
			switch transaction.TransactionType {
			case "deposit":
				account.Balance += transaction.Amount
			case "withdrawal":
				if account.Balance < transaction.Amount {
					return gorm.ErrInvalidData
				}
				account.Balance -= transaction.Amount
			case "transfer", "payment":
				if account.Balance < transaction.Amount {
					return gorm.ErrInvalidData
				}
				account.Balance -= transaction.Amount
			}

			// Update balance after transaction
			transaction.BalanceAfter = account.Balance
			transaction.TransactionID = generateTransactionID()

			// Update account balance
			if err := tx.Save(&account).Error; err != nil {
				return err
			}

			// Create transaction record
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
				return
			}
			if err == gorm.ErrInvalidData {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance or invalid account status"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transaction"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Transaction processed successfully",
			"transaction": transaction,
		})
	}
}

// GetTransactions retrieves all transactions with filtering options
func GetTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		var transactions []models.Transaction
		query := db.Preload("Account.Customer")

		// Optional filtering by account ID
		if accountID := c.Query("account_id"); accountID != "" {
			if id, err := strconv.ParseUint(accountID, 10, 32); err == nil {
				query = query.Where("account_id = ?", uint(id))
			}
		}

		// Optional filtering by transaction type
		if transactionType := c.Query("type"); transactionType != "" {
			query = query.Where("transaction_type = ?", transactionType)
		}

		var total int64
		query.Model(&models.Transaction{}).Count(&total)
		
		err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&transactions).Error
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"total":        total,
			"page":         page,
			"limit":        limit,
		})
	}
}

// ==================== LOAN HANDLERS ====================

// CreateLoan creates a new loan for a customer
// Core banking function - loan origination
func CreateLoan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loan models.Loan
		
		if err := c.ShouldBindJSON(&loan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Validate customer exists
		var customer models.Customer
		if err := db.First(&customer, loan.CustomerID).Error; err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}

		// Validate loan parameters
		if loan.PrincipalAmount <= 0 || loan.InterestRate <= 0 || loan.LoanTerm <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan parameters"})
			return
		}

		// Calculate monthly payment using standard amortization formula
		// M = P * [r(1+r)^n] / [(1+r)^n - 1]
		monthlyRate := loan.InterestRate / 12 // Convert annual rate to monthly
		power := 1.0 // Changed to float64 for proper calculation
		for i := 0; i < loan.LoanTerm; i++ {
			power *= (1 + monthlyRate)
		}
		
		numerator := loan.PrincipalAmount * monthlyRate * float64(power)
		denominator := float64(power) - 1
		loan.MonthlyPayment = numerator / denominator

		// Set loan properties
		loan.LoanNumber = generateLoanNumber()
		loan.RemainingBalance = loan.PrincipalAmount
		loan.Status = "active"
		loan.DisbursementDate = time.Now().Format("2006-01-02")
		loan.DueDate = time.Now().AddDate(0, loan.LoanTerm, 0).Format("2006-01-02")

		// Create automatic payment account for the loan
		var loanAccount models.Account
		loanAccount.CustomerID = loan.CustomerID
		loanAccount.AccountNumber = generateAccountNumber()
		loanAccount.AccountType = "loan"
		loanAccount.Balance = -loan.PrincipalAmount // Negative balance represents debt
		loanAccount.Currency = "USD"
		loanAccount.Status = "active"

		// Create loan and associated account in transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&loan).Error; err != nil {
				return err
			}
			if err := tx.Create(&loanAccount).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create loan"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Loan created successfully",
			"loan":    loan,
		})
	}
}

// GetLoans retrieves all loans with customer information
func GetLoans(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		var loans []models.Loan
		var total int64

		db.Model(&models.Loan{}).Count(&total)
		err := db.Preload("Customer").Offset(offset).Limit(limit).Find(&loans).Error
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve loans"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"loans": loans,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Additional handler stubs for completeness - similar pattern to above
func UpdateAccount(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Update account"}) } }
func DeleteAccount(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Delete account"}) } }
func UpdateLoan(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Update loan"}) } }
func DeleteLoan(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Delete loan"}) } }
func GetLoan(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Get loan"}) } }