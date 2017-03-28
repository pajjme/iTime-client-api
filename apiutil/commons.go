package apiutil

import (
	"math/rand"
	"encoding/json"
	"fmt"
	"reflect"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

// Substitute for UUID. Will be replaced.
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


// Source: https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
func AssertEqualJSON(s1, s2 string) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		panic("Error mashalling string 1 :: "+ err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		panic("Error mashalling string 2 :: %s"+ err.Error())
	}

	if !reflect.DeepEqual(o1, o2) {
		panic(fmt.Sprintf("JSON not equal: %s vs %s", s1, s2))
	}
}
