package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/videocall/getoffers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write([]byte("Hey look data"))
			fmt.Println("got the get request buddy")
		} else {
			fmt.Println("not a get request")
		}
	})

	http.HandleFunc("/videocall/postoffer", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Called")
		if r.Method == "POST" {
			body, err := io.ReadAll(r.Body)

			if err != nil {
				fmt.Println("Error")
				return
			}

			fmt.Println(string(body))

		}
	})

	fmt.Errorf("err %s", http.ListenAndServe(":42069", nil))
}
