package db

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)

	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  randomMoney(),
		Currency: randomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotNil(t, account)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	return account
}

func TestQueries_CreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestQueries_GetAccount(t *testing.T) {
	randomAcc := createRandomAccount(t)

	account, err := testQueries.GetAccount(context.Background(), randomAcc.ID)

	require.NoError(t, err)
	require.NotNil(t, account)

	require.Equal(t, randomAcc.ID, account.ID)
	require.Equal(t, randomAcc.Owner, account.Owner)
	require.Equal(t, randomAcc.Balance, account.Balance)
	require.Equal(t, randomAcc.Currency, account.Currency)
	require.WithinDuration(t, randomAcc.CreatedAt, account.CreatedAt, time.Second)
}

func TestQueries_DeleteAccount(t *testing.T) {
	randomAccount := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), randomAccount.ID)

	require.NoError(t, err)

	account, err := testQueries.GetAccount(context.Background(), randomAccount.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, account)
}

func TestQueries_ListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)
}

func getRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixMicro()))
}

func randomInt(min, max int) int {
	return min + getRand().Intn(max-min)
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[getRand().Intn(len(letter))]
	}
	return string(b)
}

func randomOwner() string {
	return randomString(6)
}

func randomMoney() int64 {
	return int64(randomInt(0, 1000))
}

func randomCurrency() string {
	currencies := []string{"USD", "EUR"}
	n := randomInt(0, len(currencies)-1)
	return currencies[n]
}
