package models

import (
	"time"
	"gorm.io/gorm"
)

// Customer represents a bank customer with basic personal information
// Core banking requires customer identification and contact details
type Customer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                    // Unique customer identifier
	CreatedAt time.Time      `json:"created_at"`                             // Record creation timestamp
	UpdatedAt time.Time      `json:"updated_at"`                             // Last update timestamp
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                         // Soft delete support
	
	// Personal Information - Essential for KYC (Know Your Customer) compliance
	FirstName  string `json:"first_name" gorm:"size:100;not null"`           // Customer's first name
	LastName   string `json:"last_name" gorm:"size:100;not null"`            // Customer's last name
	Email      string `json:"email" gorm:"size:255;uniqueIndex"`             // Unique email for identification
	Phone      string `json:"phone" gorm:"size:20"`                          // Contact phone number
	Address    string `json:"address" gorm:"size:500"`                       // Customer address
	DateOfBirth string `json:"date_of_birth" gorm:"type:date"`               // DOB for age verification
	
	// Customer Status - Important for account management
	Status string `json:"status" gorm:"size:20;default:'active'"`            // Customer status (active/inactive)
	
	// Relationships - Core banking requires linking customers to accounts and loans
	Accounts []Account `json:"accounts,omitempty"`                           // Customer's bank accounts
	Loans    []Loan    `json:"loans,omitempty"`                             // Customer's loans
}

// Account represents a bank account (checking, savings, etc.)
// Core banking systems must track account balances and types
type Account struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                   // Unique account identifier
	CreatedAt time.Time      `json:"created_at"`                            // Account creation date
	UpdatedAt time.Time      `json:"updated_at"`                            // Last update timestamp
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                        // Soft delete support
	
	// Account Identification
	AccountNumber string `json:"account_number" gorm:"size:50;uniqueIndex;not null"` // Unique account number
	CustomerID    uint   `json:"customer_id" gorm:"not null;index"`                 // Link to customer
	
	// Account Properties
	AccountType  string  `json:"account_type" gorm:"size:20;not null"`       // checking, savings, loan
	Balance      float64 `json:"balance" gorm:"type:decimal(15,2);default:0"` // Current balance
	Currency     string  `json:"currency" gorm:"size:3;default:'USD'"`       // ISO currency code
	
	// Account Status - Critical for transaction processing
	Status string `json:"status" gorm:"size:20;default:'active'"`           // Account status
	
	// Relationships
	Customer     Customer     `json:"customer,omitempty"`                    // Account owner
	Transactions []Transaction `json:"transactions,omitempty"`               // Account transaction history
}

// Transaction represents financial transactions (deposits, withdrawals, transfers)
// Core banking requires audit trail of all financial movements
type Transaction struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                   // Unique transaction ID
	CreatedAt time.Time      `json:"created_at"`                            // Transaction timestamp
	UpdatedAt time.Time      `json:"updated_at"`                            // Last update timestamp
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                        // Soft delete support
	
	// Transaction Identification
	TransactionID string `json:"transaction_id" gorm:"size:100;uniqueIndex;not null"` // System-generated transaction ID
	AccountID     uint   `json:"account_id" gorm:"not null;index"`                   // Source account
	
	// Transaction Details
	TransactionType string  `json:"transaction_type" gorm:"size:20;not null"` // deposit, withdrawal, transfer, payment
	Amount          float64 `json:"amount" gorm:"type:decimal(15,2);not null"`  // Transaction amount
	
	// Transaction Context
	Description string `json:"description" gorm:"size:500"`                   // Transaction description
	Reference   string `json:"reference" gorm:"size:100"`                     // External reference number
	
	// Balance Tracking - Critical for audit trails
	BalanceBefore float64 `json:"balance_before" gorm:"type:decimal(15,2)"`   // Balance before transaction
	BalanceAfter  float64 `json:"balance_after" gorm:"type:decimal(15,2)"`    // Balance after transaction
	
	// Relationships
	Account Account `json:"account,omitempty"`                               // Account that owns this transaction
}

// Loan represents loan products and their management
// Core banking includes loan origination and repayment tracking
type Loan struct {
	ID        uint           `json:"id" gorm:"primaryKey"`                   // Unique loan identifier
	CreatedAt time.Time      `json:"created_at"`                            // Loan creation date
	UpdatedAt time.Time      `json:"updated_at"`                            // Last update timestamp
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                        // Soft delete support
	
	// Loan Identification
	LoanNumber  string `json:"loan_number" gorm:"size:50;uniqueIndex;not null"` // Unique loan number
	CustomerID  uint   `json:"customer_id" gorm:"not null;index"`                // Link to customer
	
	// Loan Terms
	PrincipalAmount float64 `json:"principal_amount" gorm:"type:decimal(15,2);not null"` // Original loan amount
	InterestRate    float64 `json:"interest_rate" gorm:"type:decimal(5,4);not null"`     // Annual interest rate
	LoanTerm        int     `json:"loan_term" gorm:"not null"`                           // Loan term in months
	
	// Loan Status
	Status string `json:"status" gorm:"size:20;default:'active'"`           // active, paid_off, defaulted
	
	// Loan Balance Tracking
	RemainingBalance float64 `json:"remaining_balance" gorm:"type:decimal(15,2)"` // Current outstanding balance
	MonthlyPayment   float64 `json:"monthly_payment" gorm:"type:decimal(10,2)"`   // Calculated monthly payment
	
	// Dates
	DisbursementDate string `json:"disbursement_date" gorm:"type:date"`     // When loan was disbursed
	DueDate          string `json:"due_date" gorm:"type:date"`              // Final payment due date
	
	// Relationships
	Customer Customer `json:"customer,omitempty"`                           // Loan borrower
}