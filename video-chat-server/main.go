package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	postedMessages := []string{}

	// need to manage connections. I must take them, hold them, discard them, transmit them.

	http.HandleFunc("/videocall/getoffers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			for _, message := range postedMessages {

				w.Write([]byte(message))
				fmt.Println("got the get request buddy")
			}
		} else {
			fmt.Println("not a get request")
		}
	})

	http.HandleFunc("/videocall/postoffer", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Called")
		if r.Method == "POST" {
			fmt.Println("POST CALLED")
			body, err := io.ReadAll(r.Body)

			if err != nil {
				fmt.Println("Error")
				return
			}

			fmt.Println(string(body))
			postedMessages = append(postedMessages, string(body))

		}
	})

	fmt.Errorf("err %s", http.ListenAndServe(":42069", nil))
}
