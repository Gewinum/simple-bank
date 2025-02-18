package security

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"simple-bank/internal/random"
	"testing"
)

func TestComparePasswordAndHash(t *testing.T) {
	// it's "testing"
	rawPassword := "testing"
	hashedPassword := "$argon2id$v=19$m=65536,t=1,p=12$vdPiXd+rcQ+8hcs3mwrTMQ$K8PhkQpORBggghBnxZ6EGNX+6Alv6iYflUSRvTPJqIc"
	invalidHashFormat := "notevenahashlol"
	someOtherHash := "$argon2id$v=19$m=65536,t=1,p=12$wrxuS7K3poi+ooHcSt$argon2id$v=19$m=65536,t=1,p=12$wrxuS7K3poi+ooHcStNPUQ$AXTsuykut/CQecmvxXBcW57IvxyZslsWK99Y+a0iTWMNPUQ$AXTsuykut/CQecmvxXBcW57IvxyZslsWK99Y+a0iTWM"

	require.NoError(t, ComparePasswordAndHash(rawPassword, hashedPassword))
	require.Error(t, ComparePasswordAndHash(rawPassword, invalidHashFormat))
	require.Error(t, ComparePasswordAndHash(rawPassword, someOtherHash))
}

func TestHashPassword(t *testing.T) {
	rawPassword := random.String(6)
	hashedPassword, err := HashPassword(rawPassword)
	fmt.Println(hashedPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	require.NoError(t, ComparePasswordAndHash(rawPassword, hashedPassword))
}
