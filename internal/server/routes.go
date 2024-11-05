package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/frsfahd/go-expenseTracker/docs"
	"github.com/frsfahd/go-expenseTracker/internal/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("POST /register", Chain(s.RegisterHandler, Logging()))
	mux.HandleFunc("POST /login", Chain(s.LoginHandler, Logging()))

	mux.HandleFunc("POST /expenses", Chain(s.AddExpenseHandler, Auth(), Logging()))
	mux.HandleFunc("GET /expenses", Chain(s.ListExpenseHandler, Auth(), Logging()))
	mux.HandleFunc("PUT /expenses/{id}", Chain(s.UpdateExpenseHandler, Auth(), Logging()))
	mux.HandleFunc("DELETE /expenses/{id}", Chain(s.DeleteExpenseHandler, Auth(), Logging()))

	mux.Handle("/", http.FileServer(http.FS(docs.DocsFS)))

	return wrapMiddleware(mux, CORS)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// check email
	existingUser, err := s.db.Query().GetUser(context.Background(), user.Email)
	// response: 401
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		sendHTTPResponse(w, http.StatusUnauthorized, INCORRECT_EMAIL_STATUS, nil)
		return
	}

	// check password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	// response: 401
	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		sendHTTPResponse(w, http.StatusUnauthorized, INCORRECT_PWD_STATUS, nil)
		return
	}

	// response: 200
	token := signToken(existingUser)
	sendHTTPResponse(w, http.StatusOK, LOGGED_IN_STATUS, TokenData{
		Token: token,
	})

}

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// validate input
	// response: 400
	if !validateInput(user) {
		sendHTTPResponse(w, http.StatusBadRequest, BAD_REGISTER_REQ_STATUS, nil)
		return
	}

	bytes, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

	// check if email already existed
	_, err := s.db.Query().GetUser(context.Background(), user.Email)
	// response: 409
	if err == nil {
		sendHTTPResponse(w, http.StatusConflict, EMAIL_EXISTED_STATUS, nil)
		return
	}

	newUser, err := s.db.Query().AddUser(context.Background(), sqlc.AddUserParams{Email: user.Email, Password: string(bytes), Username: user.Username})

	// response: 500
	if err != nil {
		slog.Error(err.Error())
		sendHTTPResponse(w, http.StatusInternalServerError, INTERNAL_ERR_STATUS, nil)
		return
	}

	// response: 200
	sendHTTPResponse(w, http.StatusOK, USER_ADDED_STATUS, UserResponse{
		ID:       newUser.ID,
		Email:    newUser.Email,
		Username: newUser.Username,
	})
}

func (s *Server) AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
	var expense Expense
	json.NewDecoder(r.Body).Decode(&expense)

	// retrieve user credential from auth middleware
	user := r.Context().Value(USER_KEY_CTX).(LoginData)

	// input validation
	if expense.Category == "" {
		expense.Category = "general"
	}
	if expense.Desc == "" {
		expense.Desc = "-"
	}
	addExpense := sqlc.AddExpenseParams{
		Name:        expense.Name,
		Category:    expense.Category,
		Description: expense.Desc,
		Amount:      expense.Amount,
		UserID:      user.USER_ID,
	}
	// response: 400
	if !validateInput(addExpense) {
		sendHTTPResponse(w, http.StatusBadRequest, BAD_EXPENSE_REQ_STATUS, nil)
		return
	}

	updatedEx, err := s.db.Query().AddExpense(context.Background(), addExpense)

	// response: 500
	if err != nil {
		sendHTTPResponse(w, http.StatusInternalServerError, INTERNAL_ERR_STATUS, nil)
		return
	}

	// response: 201
	sendHTTPResponse(w, http.StatusCreated, EXPENSE_ADDED, ExpenseResponse{
		ID:        updatedEx.ID,
		Expense:   Expense{Name: updatedEx.Name, Desc: updatedEx.Description, Category: updatedEx.Category, Amount: updatedEx.Amount},
		CreatedAt: updatedEx.CreatedAt.Time,
		UpdatedAt: updatedEx.UpdatedAt.Time,
	})
}

func (s *Server) ListExpenseHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve user credential from auth middleware
	user := r.Context().Value(USER_KEY_CTX).(LoginData)

	var listExpenses []sqlc.Expense
	var err error

	filterTime := r.URL.Query().Get("filterTime")

	if filterTime == "fixed" {
		var filterBody TimeFilter
		json.NewDecoder(r.Body).Decode(&filterBody)

		listExpenses, err = s.db.Query().FilterExpense(context.Background(), sqlc.FilterExpenseParams{UserID: user.USER_ID, Column2: getTimeInterval(TimeInterval(filterBody.Start))})
	} else if filterTime == "custom" {
		var filterBody TimeFilter
		json.NewDecoder(r.Body).Decode(&filterBody)

		a, _ := time.Parse(time.RFC3339, filterBody.Start)
		b, _ := time.Parse(time.RFC3339, filterBody.End)
		log.Print(a, b)
		listExpenses, err = s.db.Query().FilterExpenseCustom(context.Background(), sqlc.FilterExpenseCustomParams{UserID: user.USER_ID, CreatedAt: sql.NullTime{Valid: true, Time: a}, CreatedAt_2: sql.NullTime{Valid: true, Time: b}})
	} else {
		listExpenses, err = s.db.Query().ListExpenses(context.Background(), user.USER_ID)
	}

	if err != nil {
		// response: 200 (empty list)
		if errors.Is(err, sql.ErrNoRows) {
			listExpenses = []sqlc.Expense{}
			sendHTTPResponse(w, http.StatusOK, SUCCESS, listExpenses)
			return
		}
		// response: 500
		slog.Error("error: %s", err.Error())
		sendHTTPResponse(w, http.StatusInternalServerError, INTERNAL_ERR_STATUS, nil)
		return
	}

	//response: 200 ok
	expenseRes := make([]ExpenseResponse, len(listExpenses))
	for i, ex := range listExpenses {
		expenseRes[i] = ExpenseResponse{
			ID:        ex.ID,
			Expense:   Expense{Name: ex.Name, Desc: ex.Description, Category: ex.Category, Amount: ex.Amount},
			CreatedAt: ex.CreatedAt.Time,
			UpdatedAt: ex.UpdatedAt.Time,
		}
	}
	sendHTTPResponse(w, http.StatusOK, SUCCESS, expenseRes)

}

func (s *Server) UpdateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	var expense Expense
	json.NewDecoder(r.Body).Decode(&expense)

	id, _ := strconv.Atoi(r.PathValue("id"))

	// retrieve user credential from auth middleware
	user := r.Context().Value(USER_KEY_CTX).(LoginData)

	// input validation
	if expense.Category == "" {
		expense.Category = "general"
	}
	if expense.Desc == "" {
		expense.Desc = "-"
	}

	newExpense := sqlc.UpdateExpenseParams{
		Name:        expense.Name,
		Category:    expense.Category,
		Description: expense.Desc,
		Amount:      expense.Amount,
		ID:          int32(id),
		UserID:      user.USER_ID,
	}

	// response: 400
	if !validateInput(newExpense) {
		sendHTTPResponse(w, http.StatusBadRequest, BAD_EXPENSE_REQ_STATUS, nil)
		return
	}

	updatedEx, err := s.db.Query().UpdateExpense(context.Background(), newExpense)

	if err != nil {
		// response: 404
		if errors.Is(err, sql.ErrNoRows) {
			sendHTTPResponse(w, http.StatusNotFound, NOT_FOUND, nil)
			return
		}

		// response: 500
		sendHTTPResponse(w, http.StatusInternalServerError, INTERNAL_ERR_STATUS, nil)
		return
	}

	// response: 200
	sendHTTPResponse(w, http.StatusCreated, EXPENSE_UPDATED, ExpenseResponse{
		ID:        updatedEx.ID,
		Expense:   Expense{Name: updatedEx.Name, Desc: updatedEx.Description, Category: updatedEx.Category, Amount: updatedEx.Amount},
		CreatedAt: updatedEx.CreatedAt.Time,
		UpdatedAt: updatedEx.UpdatedAt.Time,
	})
}

func (s *Server) DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve user credential from auth middleware
	user := r.Context().Value(USER_KEY_CTX).(LoginData)

	id, _ := strconv.Atoi(r.PathValue("id"))

	deletedEntry, err := s.db.Query().DeleteExpense(context.Background(), sqlc.DeleteExpenseParams{ID: int32(id), UserID: user.USER_ID})

	if err != nil {
		// response: 404
		if errors.Is(err, sql.ErrNoRows) {
			sendHTTPResponse(w, http.StatusNotFound, NOT_FOUND, nil)
			return
		}

		// response: 500
		sendHTTPResponse(w, http.StatusInternalServerError, INTERNAL_ERR_STATUS, nil)
		return
	}

	// response: 200
	sendHTTPResponse(w, http.StatusOK, SUCCESS, deletedEntry)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
