package main

import "fmt"

func fibonacci(n int) int {
	// Recursively calculate the nth fibonacci number.
	if n < 2 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	for i := 0; i < 10; i++ {
		fmt.Printf("fibonacci(%d) = %d\n", i, fibonacci(i))
	}
}
