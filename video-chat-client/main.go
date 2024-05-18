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
	makeofferFile, err := os.Open("assets/js/makeoffer.js")
	if err != nil {
		fmt.Println(err)
		return
	}
	makeofferInfo, err := makeofferFile.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	makeofferFileBytes := make([]byte, makeofferInfo.Size())

	makeofferFile.Read(makeofferFileBytes)

	fmt.Println("Need to connect to server.")

	apiURL := "http://127.0.0.1:4009"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		resp, err := http.Get(apiURL + "/videocall/getoffers")
		must(err)

		responseBody, err := io.ReadAll(resp.Body)
		must(err)

		w.Write(responseBody)
		resp.Body.Close()

	})

	http.HandleFunc("/videocall/makeoffer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
			w.Header().Set("Access-Control-Expose-Headers", "Authorization")
			comp := MakeOffer()
			comp.Render(context.Background(), w)
		} else {
			fmt.Println("NO OTHER OPTIONS")
		}
	})

	http.HandleFunc("/videocall/assets/js/makeoffer.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(makeofferFileBytes)
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
