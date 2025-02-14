package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       randomOwner(),
		HashedPassword: "secret",
		FullName:       randomOwner(),
		Email:          randomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotNil(t, user)

	require.NotEmpty(t, user.Username)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	return user
}

func TestQueries_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_GetUser(t *testing.T) {
	randomUser := createRandomUser(t)

	account, err := testQueries.GetUser(context.Background(), randomUser.Username)

	require.NoError(t, err)
	require.NotNil(t, account)

	require.Equal(t, randomUser.Username, account.Username)
	require.Equal(t, randomUser.HashedPassword, account.HashedPassword)
	require.Equal(t, randomUser.FullName, account.FullName)
	require.Equal(t, randomUser.Email, account.Email)
	require.WithinDuration(t, randomUser.PasswordChangedAt, account.PasswordChangedAt, time.Second)
	require.WithinDuration(t, randomUser.CreatedAt, account.CreatedAt, time.Second)
}

func randomEmail() string {
	return fmt.Sprintf("%s@email.com", randomOwner())
}
