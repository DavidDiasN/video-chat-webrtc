package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnInfo struct {
	Conn websocket.Conn
	uuid string
}

var postedMessages = map[string]ConnInfo{"david": ConnInfo{websocket.Conn{}, uuid.NewString()}, "Michael": ConnInfo{websocket.Conn{}, uuid.NewString()}, "George": ConnInfo{websocket.Conn{}, uuid.NewString()}}

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
	//	w.Header().Add("Access-Control-Allow-Origin", "*")
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
		if k == cleanedInput {
			w.Write([]byte("NO"))
			return
		}
	}
	returnUUID := uuid.NewString()
	postedMessages[cleanedInput] = ConnInfo{websocket.Conn{}, returnUUID}
	w.Write([]byte(returnUUID))
}

//"/videocall/makeoffer/ws",

func videocallMakeOfferWS(w http.ResponseWriter, r *http.Request) {
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

}
