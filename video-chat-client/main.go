package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("Need to connect to server.")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		apiURL := "http://127.0.0.1:42069"

		resp, err := http.Get(apiURL + "/videocall/getoffers")
		must(err)

		responseBody, err := io.ReadAll(resp.Body)
		must(err)

		w.Write(responseBody)
		resp.Body.Close()

		dareader := bytes.NewReader([]byte("Posted"))
		resp, err = http.Post(apiURL+"/videocall/postoffer", "text/plain", dareader)
		resp.Body.Close()

	})
}

func must(err error) {
	if err != nil {
		panic(err)
	}
	return
}
