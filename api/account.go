package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"log"
	"net/http"
	"simple-bank/internal/db"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (s *Server) createAccount(c *gin.Context) {
	var request createAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getPayloadFromGinCtx(c)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Subject,
		Balance:  0,
		Currency: request.Currency,
	}

	account, err := s.store.CreateAccount(c, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr.Code.Name(), pqErr.Message)

			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				c.JSON(http.StatusForbidden, errorResponse(err))
			}
		} else {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, account)
}

type getAccountParams struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getAccount(c *gin.Context) {
	var request getAccountParams
	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := s.store.GetAccount(c, request.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := getPayloadFromGinCtx(c)

	if authPayload.Subject != account.Owner {
		c.JSON(http.StatusForbidden, errorResponse(errors.New("you do not own this account")))
		return
	}

	c.JSON(http.StatusOK, account)
}

type ListAccountParams struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (s *Server) listAccounts(c *gin.Context) {
	var request ListAccountParams
	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getPayloadFromGinCtx(c)

	arg := db.ListAccountsParams{
		Owner:  authPayload.Subject,
		Limit:  request.PageSize,
		Offset: (request.PageID - 1) * request.PageSize,
	}

	accounts, err := s.store.ListAccounts(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, accounts)
}
