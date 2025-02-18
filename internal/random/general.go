package random

import (
	"math/rand"
	"time"
)

func getRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixMicro()))
}

func Int(min, max int) int {
	return min + getRand().Intn(max-min)
}

func Int64(min, max int64) int64 {
	return min + getRand().Int63n(max-min)
}

func String(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[getRand().Intn(len(letter))]
	}
	return string(b)
}
