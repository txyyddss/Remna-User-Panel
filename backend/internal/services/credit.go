package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/models"
)

// CreditService handles all TXB credit operations
type CreditService struct{}

// NewCreditService creates a new CreditService
func NewCreditService() *CreditService {
	return &CreditService{}
}

// GetBalance gets a user's current TXB balance
func (s *CreditService) GetBalance(ctx context.Context, userID int64) (float64, error) {
	var credit float64
	err := database.DB().QueryRowContext(ctx, "SELECT credit FROM users WHERE id = ?", userID).Scan(&credit)
	return credit, err
}

// AddCredit adds/deducts credit and logs it
func (s *CreditService) AddCredit(ctx context.Context, userID int64, amount float64, reason string) (float64, error) {
	tx, err := database.DB().Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Update balance
	_, err = tx.ExecContext(ctx, "UPDATE users SET credit = credit + ?, updated_at = ? WHERE id = ?",
		amount, time.Now(), userID)
	if err != nil {
		return 0, err
	}

	// Get new balance
	var newBalance float64
	err = tx.QueryRowContext(ctx, "SELECT credit FROM users WHERE id = ?", userID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}

	// Round to 2 decimal places
	newBalance = math.Round(newBalance*100) / 100

	// Log the change
	_, err = tx.ExecContext(ctx,
		"INSERT INTO credit_logs (user_id, amount, balance, reason, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, math.Round(amount*100)/100, newBalance, reason, time.Now(),
	)
	if err != nil {
		return 0, err
	}

	return newBalance, tx.Commit()
}

// ConsumeCredit deducts credit atomically and avoids concurrent overspending.
func (s *CreditService) ConsumeCredit(ctx context.Context, userID int64, amount float64, reason string) (float64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}

	tx, err := database.DB().Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		"UPDATE users SET credit = credit - ?, updated_at = ? WHERE id = ? AND credit >= ?",
		amount, time.Now(), userID, amount,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("insufficient credit")
	}

	var newBalance float64
	err = tx.QueryRowContext(ctx, "SELECT credit FROM users WHERE id = ?", userID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}

	newBalance = math.Round(newBalance*100) / 100

	_, err = tx.ExecContext(ctx,
		"INSERT INTO credit_logs (user_id, amount, balance, reason, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, -math.Round(amount*100)/100, newBalance, reason, time.Now(),
	)
	if err != nil {
		return 0, err
	}

	return newBalance, tx.Commit()
}

// Signup performs daily signup (check-in) with weighted random reward
func (s *CreditService) Signup(ctx context.Context, userID int64) (float64, float64, error) {
	cfg := config.Get()

	// Check if credit is below zero
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	if balance < 0 {
		return 0, 0, fmt.Errorf("insufficient balance, cannot check in")
	}

	// Check if already signed up today
	today := time.Now().Format("2006-01-02")
	var count int
	err = database.DB().QueryRowContext(ctx,
		"SELECT COUNT(*) FROM signup_logs WHERE user_id = ? AND date = ?",
		userID, today,
	).Scan(&count)
	if err != nil {
		return 0, 0, err
	}
	if count > 0 {
		return 0, 0, fmt.Errorf("already checked in today")
	}

	// Generate weighted random value (exponential decay: low values more likely)
	// Using inverse exponential distribution
	minVal := cfg.Credit.SignupMin
	maxVal := cfg.Credit.SignupMax

	// u is uniform [0,1), we transform it to have exponential decay toward maxVal
	u := rand.Float64()
	// Exponential distribution: higher probability for lower values
	// value = min + (max - min) * (1 - u^2)  -- but we want LOW more likely
	value := minVal + (maxVal-minVal)*math.Pow(u, 3) // cubic: strongly favors low values
	value = math.Round(value*100) / 100              // 2 decimal places

	// Record signup
	_, err = database.DB().ExecContext(ctx,
		"INSERT INTO signup_logs (user_id, date, value, created_at) VALUES (?, ?, ?, ?)",
		userID, today, value, time.Now(),
	)
	if err != nil {
		return 0, 0, err
	}

	// Add credit
	newBalance, err := s.AddCredit(ctx, userID, value, fmt.Sprintf("daily check-in +%.2f", value))
	if err != nil {
		return 0, 0, err
	}

	return value, newBalance, nil
}

// Bet performs a betting operation with weighted probabilities
func (s *CreditService) Bet(ctx context.Context, userID int64, betAmount float64) (float64, float64, error) {
	if math.IsNaN(betAmount) || math.IsInf(betAmount, 0) || betAmount <= 0 {
		return 0, 0, fmt.Errorf("bet amount must be a finite positive number")
	}

	cfg := config.Get()
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	if balance < betAmount {
		return 0, 0, fmt.Errorf("insufficient balance")
	}

	// Calculate result range: [-betAmount*lossMultiplier, betAmount*winMultiplier]
	lossMax := betAmount * cfg.Credit.BetLossMultiplier // e.g., 3x
	winMax := betAmount * cfg.Credit.BetWinMultiplier   // e.g., 2x

	// Factors affecting probability:
	// 1. Higher balance ?lower chance of winning
	// 2. More bets this month ?lower chance of positive result

	// Get monthly bet count
	monthStart := time.Now().Format("2006-01") + "-01"
	var monthlyBets int
	database.DB().QueryRowContext(ctx,
		"SELECT COUNT(*) FROM bet_logs WHERE user_id = ? AND created_at >= ?",
		userID, monthStart,
	).Scan(&monthlyBets)

	// Balance factor: higher balance = shift toward loss
	balanceFactor := math.Min(balance/1000.0, 1.0) // cap at 1000 TXB

	// Frequency factor: more bets = shift toward loss
	freqFactor := math.Min(float64(monthlyBets)/30.0, 1.0)

	// Combined bias toward loss (0 = no bias, 1 = max bias toward loss)
	lossBias := (balanceFactor*0.5 + freqFactor*0.5)

	// Generate result using biased random
	u := rand.Float64()
	// Shift u based on loss bias: higher lossBias pushes result toward negative
	adjustedU := u * (1 - lossBias*0.7) // reduce chance of high values

	// Map to range [-lossMax, winMax]
	totalRange := lossMax + winMax
	result := -lossMax + adjustedU*totalRange
	result = math.Round(result*100) / 100

	// Deduct bet amount first
	_, err = s.ConsumeCredit(ctx, userID, betAmount, fmt.Sprintf("bet -%.2f", betAmount))
	if err != nil {
		return 0, 0, err
	}

	// Add result (can be negative, further reducing balance)
	newBalance, err := s.AddCredit(ctx, userID, result, fmt.Sprintf("bet result %+.2f", result))
	if err != nil {
		return 0, 0, err
	}

	// Record bet
	database.DB().ExecContext(ctx,
		"INSERT INTO bet_logs (user_id, bet_amount, result, created_at) VALUES (?, ?, ?, ?)",
		userID, betAmount, result, time.Now(),
	)

	return result, newBalance, nil
}

// ApplyDiscount calculates a TXB discount on a bill.
// Returns: discountRMB, consumedTXB, finalBillRMB.
func (s *CreditService) ApplyDiscount(ctx context.Context, userID int64, billRMB float64, useTXB bool, requestedDiscountRMB float64) (float64, float64, float64, error) {
	if !useTXB || billRMB <= 0 {
		return 0, 0, billRMB, nil
	}

	cfg := config.Get()
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, billRMB, err
	}

	if balance <= 0 {
		return 0, 0, billRMB, nil
	}

	if cfg.Credit.TXBToRMBRate <= 0 {
		return 0, 0, billRMB, fmt.Errorf("invalid TXB conversion rate")
	}

	maxDiscountRMB := balance / cfg.Credit.TXBToRMBRate

	// Can't discount more than the bill
	if maxDiscountRMB > billRMB {
		maxDiscountRMB = billRMB
	}

	maxDiscountRMB = math.Floor(maxDiscountRMB*100) / 100
	if maxDiscountRMB <= 0 {
		return 0, 0, billRMB, nil
	}

	discountRMB := maxDiscountRMB
	if requestedDiscountRMB > 0 {
		discountRMB = math.Min(requestedDiscountRMB, maxDiscountRMB)
		discountRMB = math.Floor(discountRMB*100) / 100
	}

	if discountRMB <= 0 {
		return 0, 0, billRMB, nil
	}

	consumedTXB := math.Round(discountRMB*cfg.Credit.TXBToRMBRate*100) / 100
	finalBill := math.Round((billRMB-discountRMB)*100) / 100

	return discountRMB, consumedTXB, finalBill, nil
}

// ConvertPaymentToCredit adds TXB for a payment
func (s *CreditService) ConvertPaymentToCredit(ctx context.Context, userID int64, amountRMB float64) (float64, error) {
	cfg := config.Get()
	txb := amountRMB * cfg.Credit.RMBToTXBRate
	txb = math.Round(txb*100) / 100

	newBalance, err := s.AddCredit(ctx, userID, txb, fmt.Sprintf("purchase bonus +%.2f (spent %.2f RMB)", txb, amountRMB))
	return newBalance, err
}

// GetHistory gets credit history for a user
func (s *CreditService) GetHistory(ctx context.Context, userID int64, limit, offset int) ([]models.CreditLog, error) {
	rows, err := database.DB().QueryContext(ctx,
		"SELECT id, user_id, amount, balance, reason, created_at FROM credit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.CreditLog
	for rows.Next() {
		var log models.CreditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Amount, &log.Balance, &log.Reason, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("credit: iterate history: %w", err)
	}
	return logs, nil
}

// CleanupOldLogs deletes credit logs older than retention period
func (s *CreditService) CleanupOldLogs(ctx context.Context) error {
	cfg := config.Get()
	cutoff := time.Now().AddDate(0, 0, -cfg.Credit.LogRetentionDays)
	result, err := database.DB().ExecContext(ctx, "DELETE FROM credit_logs WHERE created_at < ?", cutoff)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		slog.Info("credit: cleaned up old log entries", "count", affected)
	}
	return nil
}

// Ensure sql import is used
var _ = sql.ErrNoRows
