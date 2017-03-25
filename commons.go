package main

import (
	"math/rand"
	"log"
	"encoding/json"
	"fmt"
	"reflect"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Substitute for UUID. Will be replaced.
func randomString(length int) string {
	randInt := func(min int, max int) int {
		return min + rand.Intn(max - min)
	}

	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
