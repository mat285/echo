package main

import (
	"fmt"
	"time"
)

func main() {

	num := 30

	for i := 0; i < num; i++ {
		fmt.Println(i)
		time.Sleep(time.Second)
		if i > num/2 {
			panic("failing now")
		}
	}
	fmt.Println(num)
	fmt.Println("Done :D")
}
