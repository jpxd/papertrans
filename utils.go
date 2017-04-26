package main

import (
	"fmt"
)

func check(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		panic(err.Error())
	}
}
