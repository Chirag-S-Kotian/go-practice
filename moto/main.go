package main

import "fmt"

func main() {
	// var a int
	// fmt.Println("Enter a number")
	// fmt.Scan(&a)
	// var b int
	// fmt.Println("Enter a number")
	// fmt.Scan(&b)
	// fmt.Println(a+b)
	// fmt.Println(a-b)
	// fmt.Println(a*b)
	// fmt.Println(a/b)
	// fmt.Println(a%b)

	for i := 1; i < 10; i++ {
		//fmt.Println(" + ")
		for j := 1; j < i; j++ {
			fmt.Print(" * ")
		}
	}
}
