package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	postedMessages := map[string]string{}

	// need to manage connections. I must take them, hold them, discard them, transmit them.

	http.HandleFunc("/videocall/getoffers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			for _, message := range postedMessages {

				w.Write([]byte(message))
				w.Write([]byte("\n"))
			}
		} else {
			fmt.Println("not a get request")
		}
	})

	http.HandleFunc("/videocall/answeroffer", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {
			fmt.Println("You have reached this spot")

			responseBody, err := io.ReadAll(r.Body)
			must(err)
			returnAddr := postedMessages[string(responseBody)]
			fmt.Println("Return addr: " + returnAddr)
			if returnAddr == "" {
				return
			}
			resp, err := http.Get("http://" + returnAddr + "/videocall/incomingAnswers")
			if err != nil {
				log.Fatal(err)
			}
			must(err)

			fmt.Println("How did that go")

			responseBody, err = io.ReadAll(resp.Body)
			must(err)
			fmt.Println(resp.Body)
		}

	})

	http.HandleFunc("/videocall/postoffer", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Called")
		fmt.Println(r.RemoteAddr)

		if r.Method == "POST" {
			fmt.Println("POST CALLED")
			body, err := io.ReadAll(r.Body)

			must(err)

			offer := strings.TrimPrefix(string(body), "user-input=")
			fmt.Println(offer)
			realAddr := trimToColon(r.RemoteAddr)
			postedMessages[offer] = realAddr + ":32069"

		}
	})

	fmt.Errorf("err %s", http.ListenAndServe(":42069", nil))
}

func must(err error) {

	if err != nil {
		fmt.Print("See ya")
		panic(err)
	}
}

func trimToColon(addr string) string {
	l := len(addr) - 1
	for i := l; i > 0; i-- {
		if addr[i] == ':' {
			return addr[0:i]
		}
	}
	return ""
}
