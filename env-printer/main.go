package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {

	// TODO: create a simple webserver that responds as the value of the NAME environment variable
	// Deploy this to k8s cluster, use service nodeport to expose it to local machine
	// see the result on browser

	// change the value inside configmap --> see if the pods gets reloaded with new value
	fmt.Println("Hello, World!")
	fmt.Printf("Value for env: NAME is %s\n", os.Getenv("NAME"))
	http.ListenAndServe(":8080", nil)
}
