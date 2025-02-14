package db

import (
	"context"
	"database/sql"
)

type Store interface {
	TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error)
	Querier
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (s *SQLStore) TransferTx(ctx context.Context, args TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: args.FromAccountID,
			ToAccountID:   args.ToAccountID,
			Amount:        args.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.FromAccountID,
			Amount:    -args.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.ToAccountID,
			Amount:    args.Amount,
		})
		if err != nil {
			return err
		}

		if args.FromAccountID < args.ToAccountID {
			result.FromAccount, result.ToAccount, err = s.addBalance(ctx, args.FromAccountID, -args.Amount, args.ToAccountID, args.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = s.addBalance(ctx, args.ToAccountID, args.Amount, args.FromAccountID, -args.Amount)
		}

		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func (s *SQLStore) addBalance(
	ctx context.Context,
	account1Id int64,
	amount1 int64,
	account2Id int64,
	amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = s.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account1Id,
		Amount: amount1,
	})
	if err != nil {
		return
	}
	account2, err = s.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     account2Id,
		Amount: amount2,
	})
	return
}

func (s *SQLStore) execTx(ctx context.Context, cb func(q *Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = cb(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
