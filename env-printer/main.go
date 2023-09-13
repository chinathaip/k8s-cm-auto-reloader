package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, World!")
	fmt.Printf("Value for env: NAME is %s\n", os.Getenv("NAME"))
}
