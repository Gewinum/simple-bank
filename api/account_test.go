package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"simple-bank/internal/db"
	mockdb "simple-bank/internal/db/mock"
	"simple-bank/random"
	"simple-bank/tokens"
	"testing"
	"time"
)

func TestServer_CreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		body          createAccountRequest
		setupAuth     func(t *testing.T, request *http.Request, tokensManager tokens.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			body: createAccountRequest{
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				dbRequest := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(dbRequest)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "invalid_currency",
			body: createAccountRequest{
				Currency: "not_a_currency",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "unexistent_user",
			body: createAccountRequest{
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				dbRequest := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(dbRequest)).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: "23503"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "duplicate_user_and_currency",
			body: createAccountRequest{
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				dbRequest := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(dbRequest)).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			body := tc.body
			bodyBytes, err := json.Marshal(body)
			require.NoError(t, err)
			bodyReader := bytes.NewReader(bodyBytes)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bodyReader)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokensManager)

			server.engine.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestServer_GetAccount(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, request *http.Request, tokensManager tokens.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "Internal Server Error",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "Bad Request",
			accountId: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
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
			name:      "wrong_user",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   "wrong_user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokensManager)

			server.engine.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestServer_ListAccounts(t *testing.T) {
	user, _ := randomUser(t)
	n := 10
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(user.Username)
	}

	testCases := []struct {
		name          string
		body          ListAccountParams
		setupAuth     func(t *testing.T, request *http.Request, tokensManager tokens.Manager)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			body: ListAccountParams{
				PageID:   1,
				PageSize: 10,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				dbRequest := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  10,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(dbRequest)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "empty list",
			body: ListAccountParams{
				PageID:   1,
				PageSize: 10,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				dbRequest := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  10,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(dbRequest)).
					Times(1).
					Return([]db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, []db.Account{})
			},
		},
		{
			name: "invalid_page_id",
			body: ListAccountParams{
				PageID:   0,
				PageSize: 10,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "small_page_size",
			body: ListAccountParams{
				PageID:   1,
				PageSize: 4,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "big_page_size",
			body: ListAccountParams{
				PageID:   1,
				PageSize: 11,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokensManager tokens.Manager) {
				addAuthorization(t, request, tokensManager, authorizationTypeBearer, tokens.PayloadCreationParams{
					Subject:   user.Username,
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.body.PageID, tc.body.PageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokensManager)

			server.engine.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var actualAccounts []db.Account
	err = json.Unmarshal(data, &actualAccounts)
	require.NoError(t, err)
	require.Equal(t, len(accounts), len(actualAccounts))
	for i := range actualAccounts {
		require.Equal(t, actualAccounts[i], accounts[i])
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       random.Int64(1, 1000),
		Owner:    owner,
		Balance:  random.AccountBalance(),
		Currency: random.AccountCurrency(),
	}
}
