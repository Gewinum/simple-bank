package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"simple-bank/internal/db"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransfer(c *gin.Context) {
	var request transferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := s.validateAccount(c, request.FromAccountID, request.Currency)
	if !valid {
		return
	}

	authPayload := getPayloadFromGinCtx(c)
	if authPayload.Subject != fromAccount.Owner {
		c.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("you do not own account %d", request.FromAccountID)))
		return
	}

	_, valid = s.validateAccount(c, request.ToAccountID, request.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: request.FromAccountID,
		ToAccountID:   request.ToAccountID,
		Amount:        request.Amount,
	}

	txResult, err := s.store.TransferTx(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, txResult)
}

func (s *Server) validateAccount(c *gin.Context, accountId int64, currency string) (db.Account, bool) {
	account, err := s.store.GetAccount(c, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("account [%d] currency mismatched. expected: %s, actual: %s", accountId, currency, account.Currency)))
		return account, false
	}

	return account, true
}
