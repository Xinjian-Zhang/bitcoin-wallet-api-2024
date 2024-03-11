package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Transaction represents a transaction record
type Transaction struct {
	ID        string    `json:"id" gorm:"primary_key"` // Transaction ID
	Amount    float64   `json:"amount"`                // Transaction amount
	Spent     bool      `json:"spent"`                 // Whether the transaction is spent
	CreatedAt time.Time `json:"created_at"`            // Transaction creation time
}

var db *gorm.DB

// initDB initializes the database connection
func initDB() {
	var err error
	// Use SQLite database
	db, err = gorm.Open(sqlite.Open("transactions.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Automatically migrate the database schema
	db.AutoMigrate(&Transaction{})
}

func main() {
	initDB()

	router := gin.Default()

	// Define routes
	router.GET("/transactions", listTransactions)
	router.GET("/balance", showBalance)
	router.POST("/transfer", createTransfer)

	// Start HTTP server
	router.Run(":8080")
}

// listTransactions returns all unspent transactions
func listTransactions(c *gin.Context) {
	var transactions []Transaction
	// Query unspent transactions from the database
	result := db.Where("spent = ?", false).Find(&transactions)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query transactions"})
		return
	}
	// Return the query result
	c.JSON(http.StatusOK, transactions)
}

// getBtcToEurRate retrieves the BTC to EUR exchange rate
func getBtcToEurRate() (float64, error) {
	resp, err := http.Get("http://api-cryptopia.adca.sh/v1/prices/ticker")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			Symbol string `json:"symbol"`
			Value  string `json:"value"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	for _, item := range response.Data {
		if item.Symbol == "BTC/EUR" {
			rate, err := strconv.ParseFloat(item.Value, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to convert BTC/EUR rate to float: %v", err)
			}
			return rate, nil
		}
	}

	return 0, errors.New("BTC/EUR rate not found")
}

// showBalance displays the balance
func showBalance(c *gin.Context) {
	var transactions []Transaction
	// Query unspent transactions
	db.Where("spent = ?", false).Find(&transactions)

	totalBtc := 0.0
	for _, t := range transactions {
		totalBtc += t.Amount
	}

	rate, err := getBtcToEurRate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BTC to EUR rate"})
		return
	}

	// Return balance information
	c.JSON(http.StatusOK, gin.H{
		"balance_btc": totalBtc,
		"balance_eur": totalBtc * rate,
	})
}

// createTransfer creates a transfer
func createTransfer(c *gin.Context) {
	var request struct {
		AmountEur float64 `json:"amount_eur"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rate, err := getBtcToEurRate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BTC to EUR rate"})
		return
	}

	amountBtc := request.AmountEur / rate

	if amountBtc < 0.00001 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer amount too small"})
		return
	}

	var transactions []Transaction
	db.Where("spent = ?", false).Order("created_at asc").Find(&transactions)

	totalBtc := 0.0
	var spentTransactions []string

	for _, t := range transactions {
		totalBtc += t.Amount
		spentTransactions = append(spentTransactions, t.ID)

		if totalBtc >= amountBtc {
			break
		}
	}

	if totalBtc < amountBtc {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Mark transactions as spent and handle leftover amount
	db.Model(&Transaction{}).Where("id IN ?", spentTransactions).Update("spent", true)
	if totalBtc > amountBtc {
		leftover := totalBtc - amountBtc
		newTransaction := Transaction{
			ID:        generateRandomHex(), // Generate a random transaction ID
			Amount:    leftover,
			Spent:     false,
			CreatedAt: time.Now(),
		}
		db.Create(&newTransaction)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// generateRandomHex generates a random hexadecimal string as a transaction ID
func generateRandomHex() string {
	b := make([]byte, 16) // Generate a 32-character hexadecimal string
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
