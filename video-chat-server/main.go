package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

var postedMessages = map[string]ConnInfo{"David": ConnInfo{websocket.Conn{}, uuid.NewString()}, "Michael": ConnInfo{websocket.Conn{}, uuid.NewString()}, "George": ConnInfo{websocket.Conn{}, uuid.NewString()}}

//postedMessages := map[string]websocket.Conn{"david": websocket.Conn{}, "Michael": websocket.Conn{}, "George": websocket.Conn{}}

func main() {

	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Default to the current directory")
	flag.Parse()
	router := mux.NewRouter()
	router.PathPrefix("/static/assets/js").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	router.HandleFunc("/videocall/getoffers", videocallGetOffers)
	router.HandleFunc("/videocall/MakeOffer", videocallMakeOffer)
	router.HandleFunc("/videocall/offernameValidation", videocallOfferNameValidation).Methods("POST")
	router.HandleFunc("/videocall/MakeOffer/ws", videocallMakeOfferWS)

	srv := &http.Server{
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"HEAD", "POST", "OPTIONS", "GET", "PUT"}),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token"}),
			//	handlers.AllowCredentials(),
			//handlers.IgnoreOptions()
		)(router),
		Addr:         os.Getenv("WEB_SERVER_ADDRESS"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func must(err error) {

	if err != nil {
		fmt.Print("See ya")
		panic(err)
	}
}

// /videocall/getoffers"

func videocallGetOffers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		keys := []string{}
		for k := range postedMessages {
			keys = append(keys, k)
		}
		getPage := getOffersPage(keys)
		getPage.Render(context.Background(), w)

	} else {

		conn, err := upgrader.Upgrade(w, r, nil)
		must(err)
		testMessage := []byte("<li>Anotherone</li>")
		err = conn.WriteMessage(1, testMessage)

	}

}

// "/videocall/makeoffer"

func videocallMakeOffer(w http.ResponseWriter, r *http.Request) {

	comp := makeoffer()
	comp.Render(context.Background(), w)

}

//"/videocall/offernameValidation"

func videocallOfferNameValidation(w http.ResponseWriter, r *http.Request) {

	//	w.Header().Add("Access-Control-Allow-Origin", "*")

	responseBody, err := io.ReadAll(r.Body)
	must(err)
	cleanedInput := string(responseBody)
	for k := range postedMessages {
		fmt.Println(k)
		if k == cleanedInput {
			w.Write([]byte("NO"))
			return
		}
	}
	returnUUID := uuid.NewString()
	postedMessages[cleanedInput] = ConnInfo{websocket.Conn{}, returnUUID}
	w.Write([]byte(returnUUID))
}

func validateNameHash(nameHash string) bool {
	objects := strings.Split(nameHash, " ")
	response := postedMessages[objects[0]]
	if response.uuid == "" {
		return false
	} else if response.uuid == objects[1] {
		return true
	} else {
		return false
	}

}

//"/videocall/makeoffer/ws",

func videocallMakeOfferWS(w http.ResponseWriter, r *http.Request) {
	//

	conn, err := upgrader.Upgrade(w, r, nil)
	must(err)
	var message string
	_, readBuffer, err := conn.ReadMessage()
	must(err)
	if !validateNameHash(string(readBuffer)) {
		fmt.Println("Invalid")
		return
	}

	fmt.Println("made it past")
	for message != "DONE" {
		_, readBuffer, err = conn.ReadMessage()
		must(err)
		message = string(readBuffer)
		fmt.Println(message)

	}
	fmt.Println("left")

}
