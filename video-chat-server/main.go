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

type Client struct {
	inbox chan string
	uuid  string
}

var emptyClient = Client{}

// treat all connections/clients the same. This will make things easier. the functions called will make it easy to differentiate between
// a poster or an answerer.
var answerClients = map[string]Client{}

var offerClients = map[string]Client{"David": {make(chan string), uuid.NewString()}, "Michael": Client{make(chan string), uuid.NewString()}, "George": Client{make(chan string), uuid.NewString()}}

//client:= map[string]websocket.Conn{"david": websocket.Conn{}, "Michael": websocket.Conn{}, "George": websocket.Conn{}}

func main() {

	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Default to the current directory")
	flag.Parse()
	router := mux.NewRouter()
	router.PathPrefix("/static/assets/js").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	router.HandleFunc("/videocall/MakeAnswer", videocallGetOffers)
	router.HandleFunc("/videocall/MakeOffer", videocallMakeOffer)
	router.HandleFunc("/videocall/OfferValidation", offerValidation).Methods("POST")
	router.HandleFunc("/videocall/AnswerValidation", answerValidation).Methods("POST")
	router.HandleFunc("/videocall/MakeOffer/ws", videocallOfferWS)
	router.HandleFunc("/videocall/MakeAnswer/ws", videocallAnswerWS)

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

func videocallGetOffers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		keys := []string{}
		for k := range offerClients {
			keys = append(keys, k)
		}
		getPage := getOffersPage()
		getPage.Render(context.Background(), w)

	}
}

func videocallMakeOffer(w http.ResponseWriter, r *http.Request) {

	comp := makeoffer()
	comp.Render(context.Background(), w)

}

func answerValidation(w http.ResponseWriter, r *http.Request) {
	responseBody, err := io.ReadAll(r.Body)
	must(err)
	// add some input validation or cleaning so that only letters can be used.
	cleanedInput := string(responseBody)
	searchResult := answerClients[cleanedInput]

	if searchResult != emptyClient {
		w.Write([]byte("NO"))
		return
	}

	returnUUID := uuid.NewString()
	answerClients[cleanedInput] = Client{make(chan string), returnUUID}
	w.Write([]byte(returnUUID))
}

func offerValidation(w http.ResponseWriter, r *http.Request) {
	responseBody, err := io.ReadAll(r.Body)
	must(err)
	// add some input validation or cleaning so that only letters can be used.
	cleanedInput := string(responseBody)
	searchResult := offerClients[cleanedInput]

	if searchResult != emptyClient {
		w.Write([]byte("NO"))
		return
	}

	returnUUID := uuid.NewString()
	offerClients[cleanedInput] = Client{make(chan string), returnUUID}
	w.Write([]byte(returnUUID))
}

func videocallOfferWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	must(err)
	var message string
	_, readBuffer, err := conn.ReadMessage()
	must(err)
	message = string(readBuffer)
	name, hash, boolres := nameHashOfferValidation(message)

	if !boolres {
		conn.WriteMessage(1, []byte("invalid offer"))
		return
	}

	fmt.Println(hash)

	incomingOffers := map[string]string{}

	go func() {
		for {

			select {
			case a := <-offerClients[name].inbox:
				// process the string you get
				// need to save the sdps in a map
				// maybe just in a map over on the client side tho
				splitMessage := strings.SplitN(a, " ", 2)
				fmt.Println(len(splitMessage))

				incomingOffers[splitMessage[0]] = splitMessage[1]
				fmt.Println("DONE SPLITING THE MESSAGE")
				conn.WriteMessage(1, []byte(liMaker(splitMessage[0])))
			}
		}
	}()

	for message != "DONE" {
		_, readBuffer, err = conn.ReadMessage()
		must(err)
		message = string(readBuffer)
		if message != "DONE" {
			fmt.Println(message)
			if strings.HasPrefix(message, "request") {
				message = strings.TrimPrefix(message, "request{ ")
				fmt.Println(message)

				return
				// message must contain the target and the answer.
				/*
					offerClients[message].inbox <- name
				*/
			}
		}
		// send over the stuff
		// respond with answer

		//select {
		//case a := <-answerClients[name].inbox:
		// this will be the sdp or whatever offer
		//}
		// THE SIGNAL SERVERS JOB IS DONE

	}

	// listen here for messages on the channel

	// when you get a message on the channel you need to send it to the websocket connection
	// send it in a form it can be displayed in.

}

func videocallAnswerWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	must(err)
	var message string
	_, readBuffer, err := conn.ReadMessage()
	must(err)
	message = string(readBuffer)
	name, hash, boolres := nameHashAnswerValidation(message)

	must(err)
	if !boolres {
		fmt.Println("Invalid")
		return
	}
	fmt.Println(name)
	fmt.Println(hash)

	// keep reading the offers
	lastUpdateLen := 0

	go func() {

		for {
			accu := ""
			if len(offerClients) != lastUpdateLen {
				lastUpdateLen = len(offerClients)
				for key := range offerClients {
					accu += liMaker(key)
				}
				conn.WriteMessage(1, []byte(accu))
			}
		}
	}()

	for message != "DONE" {
		_, readBuffer, err = conn.ReadMessage()
		must(err)
		message = string(readBuffer)
		if message != "DONE" {

			if strings.HasPrefix(message, "request") {

				sendTo, sdpOffer := breakDownRequest(message)

				message := fmt.Sprint(name + " " + sdpOffer)
				fmt.Println(sendTo + " " + name + " " + sdpOffer)
				fmt.Println("sending")
				offerClients[sendTo].inbox <- message
				fmt.Println("Waiting for response")

				select {
				case a := <-answerClients[name].inbox:
					// idk if this is even necessary. Maybe it will just signal that they said yes, idk
					fmt.Println(a)
					// this will be the sdp or whatever offer

				case <-time.After(60 * time.Second):
					fmt.Println("Timeout")
					return
				}

			}
		}

	}

	fmt.Println("closing")
}

func liMaker(name string) string {
	return fmt.Sprintf("<li id=\"offer-item\" onclick=\"clickName('%s')\">%s</li>", name, name)
}

func nameHashAnswerValidation(nameHash string) (string, string, bool) {
	objects := strings.Split(nameHash, " ")
	response := answerClients[objects[0]]
	if response.uuid == "" {
		return "", "", false
	} else if response.uuid == objects[1] {
		return objects[0], objects[1], true
	} else {
		return "", "", false
	}

}

func nameHashOfferValidation(nameHash string) (string, string, bool) {
	objects := strings.Split(nameHash, " ")
	response := offerClients[objects[0]]
	if response.uuid == "" {
		return "", "", false
	} else if response.uuid == objects[1] {
		return objects[0], objects[1], true
	} else {
		return "", "", false
	}

}

func breakDownRequest(requestString string) (string, string) {
	fmt.Println("START OF BREAKDOWN REQUEST")

	message := strings.TrimPrefix(requestString, "request{ ")
	fmt.Println(message)

	splitStrings := strings.Split(message, " sdp: ")
	splitStrings[0] = strings.TrimPrefix(splitStrings[0], "name: ")
	fmt.Println("end of break down request")
	fmt.Println(splitStrings[0])
	fmt.Println(splitStrings[1])
	return splitStrings[0], splitStrings[1]

}
