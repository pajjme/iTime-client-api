package api

import (
	"math/rand"
	"log"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func RandomString(length int) string {
	randInt := func(min int, max int) int {
		return min + rand.Intn(max - min)
	}

	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}
