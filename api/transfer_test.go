package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"simple-bank/internal/db"
	mockdb "simple-bank/internal/db/mock"
	"simple-bank/utils"
	"testing"
	"time"
)

func TestServer_createTransfer(t *testing.T) {
	account1 := randomAccount()
	account2 := randomAccount()

	testCases := []struct {
		name          string
		createBody    func() transferRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				requestTransfer := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
				}
				responseTransfer := db.TransferTxResult{
					Transfer: db.Transfer{
						ID:            1,
						FromAccountID: account1.ID,
						ToAccountID:   account2.ID,
						Amount:        10,
						CreatedAt:     time.Now(),
					},
					FromAccount: db.Account{ID: account1.ID},
					ToAccount:   db.Account{ID: account2.ID},
					FromEntry:   db.Entry{ID: 1, AccountID: account1.ID, Amount: -10},
					ToEntry:     db.Entry{ID: 2, AccountID: account2.ID, Amount: 10},
				}

				account1.Currency = utils.CurrencyUSD
				account2.Currency = utils.CurrencyUSD

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(requestTransfer)).
					Times(1).
					Return(responseTransfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var result db.TransferTxResult
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
			},
		},
		{
			name: "mismatched_currency_of_source",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				account1.Currency = utils.CurrencyCAD

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "mismatched_currency_of_destination",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				account1.Currency = utils.CurrencyUSD
				account2.Currency = utils.CurrencyCAD

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "unexistent_source_account",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "unexistent_destination_account",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				account1.Currency = utils.CurrencyUSD

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "no_from_account_id_in_request",
			createBody: func() transferRequest {
				return transferRequest{
					ToAccountID: account2.ID,
					Amount:      10,
					Currency:    utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "no_to_account_id_in_request",
			createBody: func() transferRequest {
				return transferRequest{
					ToAccountID: account1.ID,
					Amount:      10,
					Currency:    utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid_amount_in_request",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        0,
					Currency:      utils.CurrencyUSD,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid_currency_in_request",
			createBody: func() transferRequest {
				return transferRequest{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        10,
					Currency:      "some",
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			trReq := tc.createBody()
			rawBody, err := json.Marshal(trReq)
			require.NoError(t, err)
			require.NotEmpty(t, rawBody)

			byteBuffer := bytes.NewBuffer(rawBody)
			request, err := http.NewRequest(http.MethodPost, "/transfers", byteBuffer)
			require.NoError(t, err)

			server.engine.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
