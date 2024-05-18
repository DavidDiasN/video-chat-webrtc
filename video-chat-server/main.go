package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnInfo struct {
	Conn websocket.Conn
	uuid string
}

func main() {

	//postedMessages := map[string]websocket.Conn{"david": websocket.Conn{}, "Michael": websocket.Conn{}, "George": websocket.Conn{}}
	postedMessages := map[string]ConnInfo{"david": ConnInfo{websocket.Conn{}, uuid.NewString()}, "Michael": ConnInfo{websocket.Conn{}, uuid.NewString()}, "George": ConnInfo{websocket.Conn{}, uuid.NewString()}}

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

	getoffersFile, err := os.Open("assets/js/getoffers.js")
	if err != nil {
		fmt.Println(err)
		return
	}
	getoffersInfo, err := getoffersFile.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	getoffersFileBytes := make([]byte, getoffersInfo.Size())

	getoffersFile.Read(getoffersFileBytes)

	// need to manage connections. I must take them, hold them, discard them, transmit them.

	http.HandleFunc("/videocall/getoffers", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			keys := []string{}
			for k := range postedMessages {
				keys = append(keys, k)
			}
			getPage := getOffersPage(keys)
			getPage.Render(context.Background(), w)

		} else {
			fmt.Println("not a get request")
		}

	})

	http.HandleFunc("/videocall/makeoffer", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		comp := makeoffer()
		comp.Render(context.Background(), w)

	})

	http.HandleFunc("/videocall/offernameValidation", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		responseBody, err := io.ReadAll(r.Body)
		must(err)
		cleanedInput := string(responseBody)
		for k := range postedMessages {
			if k == cleanedInput {
				w.Write([]byte("NO"))
				return
			}
		}
		returnUUID := uuid.NewString()
		postedMessages[cleanedInput] = ConnInfo{websocket.Conn{}, returnUUID}
		w.Write([]byte(returnUUID))
	})

	http.HandleFunc("/videocall/makeoffer/ws", func(w http.ResponseWriter, r *http.Request) {
		//check body of request for name and uuid
		requestMessage, err := io.ReadAll(r.Body)
		must(err)
		fmt.Println(string(requestMessage))
		if len(requestMessage) > 2 {
			fmt.Println("has size 2 or more ")
		}
		//conn, err := upgrader.Upgrade(w, r, nil)
		//must(err)
		//conn.WriteMessage(1, []byte("Hello my friend"))

	})

	http.HandleFunc("/videocall/assets/js/makeoffer.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(makeofferFileBytes)
	})

	http.HandleFunc("/videocall/assets/js/getoffers.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(getoffersFileBytes)
	})

	fmt.Errorf("err %s", http.ListenAndServe(":4009", nil))
}

func must(err error) {

	if err != nil {
		fmt.Print("See ya")
		panic(err)
	}
}
