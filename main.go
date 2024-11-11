package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Usage: url")
		return
	}
	url := arguments[1]

	var resp *http.Response
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	fmt.Printf("%s\n", body)
}