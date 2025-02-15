package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"simple-bank/internal/db"
	mockdb "simple-bank/internal/db/mock"
	"simple-bank/random"
	"simple-bank/security"
	"testing"
)

func TestServer_createUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		createBody    func() createUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: user.Username,
					Password: password,
					FullName: user.FullName,
					Email:    user.Email,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), getCreateUserParamMatcher(user, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var result createUserResponse
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
				require.Equal(t, user.Username, result.Username)
				require.Equal(t, user.FullName, result.FullName)
				require.Equal(t, user.Email, result.Email)
			},
		},
		{
			name: "duplicate_name_or_email",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: user.Username,
					Password: password,
					FullName: user.FullName,
					Email:    user.Email,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), getCreateUserParamMatcher(user, password)).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "internal_server_error",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: user.Username,
					Password: password,
					FullName: user.FullName,
					Email:    user.Email,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), getCreateUserParamMatcher(user, password)).
					Times(1).
					Return(db.User{}, errors.New("Internal Server Error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "invalid_username",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: "username#1",
					Password: password,
					FullName: user.FullName,
					Email:    user.Email,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "invalid_email",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: user.Username,
					Password: password,
					FullName: user.FullName,
					Email:    "notvalid",
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "short_password",
			createBody: func() createUserRequest {
				return createUserRequest{
					Username: user.Username,
					Password: "short",
					FullName: user.FullName,
					Email:    user.Email,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
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
			request, err := http.NewRequest(http.MethodPost, "/users", byteBuffer)
			require.NoError(t, err)

			server.engine.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (db.User, string) {
	password := random.String(6)

	hashedPassword, err := security.HashPassword(password)
	require.NoError(t, err)

	return db.User{
		Username:       random.String(6),
		FullName:       random.String(6),
		Email:          random.UserEmail(),
		HashedPassword: hashedPassword,
	}, password
}

func getCreateUserParamMatcher(user db.User, rawPassword string) gomock.Matcher {
	return gomock.Cond[db.CreateUserParams](func(x db.CreateUserParams) bool {
		if x.Username != user.Username {
			return false
		}

		if err := security.ComparePasswordAndHash(rawPassword, x.HashedPassword); err != nil {
			return false
		}

		if x.FullName != user.FullName {
			return false
		}

		if x.Email != user.Email {
			return false
		}

		return true
	})
}
