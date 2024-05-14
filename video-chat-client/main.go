package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	htmxFile, err := os.Open("assets/js/htmx.min.js")
	if err != nil {
		fmt.Println(err)
		return
	}
	info, err := htmxFile.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	htmxFileBytes := make([]byte, info.Size())

	htmxFile.Read(htmxFileBytes)

	fmt.Println("Need to connect to server.")

	apiURL := "http://127.0.0.1:42069"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		resp, err := http.Get(apiURL + "/videocall/getoffers")
		must(err)

		responseBody, err := io.ReadAll(resp.Body)
		must(err)

		w.Write(responseBody)
		resp.Body.Close()

	})

	http.HandleFunc("/videocall/makeoffer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			responseBody, err := io.ReadAll(r.Body)
			must(err)
			responseReader := bytes.NewReader(responseBody)
			resp, err := http.Post(apiURL+"/videocall/postoffer", "text/plain", responseReader)
			must(err)
			resp.Body.Close()
			fmt.Println("Offer sent off")
		} else if r.Method == "GET" {
			comp := MakeOffer()
			comp.Render(context.Background(), w)
		} else {
			fmt.Println("NO OTHER OPTIONS")
		}
	})

	http.HandleFunc("/videocall/assets/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(htmxFileBytes)
	})

	http.HandleFunc("/videocall/incomingAnswers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("You hit the answer line")

	})

	http.HandleFunc("/videocall/makeAnswer", func(w http.ResponseWriter, r *http.Request) {
		responseReader := bytes.NewReader([]byte("David"))
		resp, err := http.Post(apiURL+"/videocall/answeroffer", "text/plain", responseReader)
		if err != nil {
			log.Fatal(err)
		}
		//must(err)
		resB, err := io.ReadAll(resp.Body)
		must(err)
		fmt.Println(string(resB))
	})

	fmt.Errorf("Err: %s", http.ListenAndServe(":32069", nil))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
	return
}
