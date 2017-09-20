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
		if i > num/2 && fail {
			panic("failing now")
		}
	}
	fmt.Println(num)
	fmt.Println("Done :D")
}
