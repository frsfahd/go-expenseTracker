package server

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type UserResponse struct {
	ID       int32  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Expense struct {
	Name     string          `json:"name"`
	Desc     string          `json:"desc"`
	Category string          `json:"category"`
	Amount   decimal.Decimal `json:"amount"`
}

type ExpenseResponse struct {
	ID int32 `json:"id"`
	Expense
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type TokenData struct {
	Token string `json:"token"`
}

type TimeFilter struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}
