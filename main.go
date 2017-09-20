package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fail := len(os.Args) > 1

	num := 30

	for i := 0; i < num; i++ {
		fmt.Println(i)
		time.Sleep(time.Second)
		if i > num/2 {
			panic("failing now")
		}
	}
	if fail {
		panic("Failing now")
	} else {
		fmt.Println(num)
		fmt.Println("Done :D")
	}
}
