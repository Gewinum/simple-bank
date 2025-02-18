package random

import (
	"fmt"
)

func Username() string {
	return String(6)
}

func UserEmail() string {
	return fmt.Sprintf("%s@example.com", Username())
}
