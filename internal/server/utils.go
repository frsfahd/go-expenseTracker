package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/frsfahd/go-expenseTracker/internal/sqlc"
	"github.com/golang-jwt/jwt/v5"
)

const (
	USER_ADDED_STATUS = "user added"
	LOGGED_IN_STATUS  = "logged in"
	EXPENSE_ADDED     = "expense added"
	EXPENSE_UPDATED   = "expense updated"
	SUCCESS           = "success"

	INCORRECT_EMAIL_STATUS  = "incorrect email"
	INCORRECT_PWD_STATUS    = "incorrect password"
	EMAIL_EXISTED_STATUS    = "email already registered, use another email"
	INTERNAL_ERR_STATUS     = "internal server error"
	BAD_REGISTER_REQ_STATUS = "all field must be valid: email, password, username"
	BAD_EXPENSE_REQ_STATUS  = "these fields must not be empty: name, amount"
	NOT_FOUND               = "cannot find the entry"
)

type TimeInterval string

const (
	PAST_WEEK     TimeInterval = "past_week"
	PAST_MONTH    TimeInterval = "past_month"
	LAST_3_MONTHS TimeInterval = "last_three_months"
)

var (
	SECRET = []byte(os.Getenv("SECRET"))
)

type LoginClaims struct {
	LoginData
	jwt.RegisteredClaims
}

type LoginData struct {
	USER_ID  int32  `json:"id"`
	Username string `json:"username"`
}

func signToken(user sqlc.User) string {

	// Create claims with multiple fields populated
	claims := LoginClaims{
		LoginData{
			USER_ID:  user.ID,
			Username: user.Username,
		},
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60000 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString(SECRET)

	return ss

}

func parseToken(tokenString string) (LoginData, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LoginClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET), nil
	})
	if err != nil {
		return LoginData{}, err
	} else if claims, ok := token.Claims.(*LoginClaims); ok {
		return LoginData{USER_ID: claims.USER_ID, Username: claims.Username}, nil
	} else {
		return LoginData{}, errors.New("unknown claims type, cannot proceed")
	}
}

func sendHTTPResponse(w http.ResponseWriter, status int, msg string, data interface{}) {
	w.WriteHeader(status)
	if data == nil {
		json.NewEncoder(w).Encode(Response{
			Message: msg,
		})
		return
	}
	json.NewEncoder(w).Encode(Response{
		Message: msg,
		Data:    data,
	})
}

func validateInput(input interface{}) bool {
	v := reflect.ValueOf(input)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			return false
		}
	}
	return true
}

func getTimeInterval(interval TimeInterval) string {
	switch interval {
	case PAST_WEEK:
		return "7 days" // 7 days in seconds
	case PAST_MONTH:
		return "30 days" // 30 days in seconds
	case LAST_3_MONTHS:
		return "3 months" // 90 days in seconds
	default:
		return "1 year"
	}
}
