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

const protocolSep string = "/:|:/"

var emptyClient = Client{}

var answerClients = map[string]Client{}

var offerClients = map[string]Client{}

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
	_, readBuffer, err := conn.ReadMessage()
	must(err)
	message := string(readBuffer)
	name, _, boolres := nameHashOfferValidation(message)

	if !boolres {
		conn.WriteMessage(1, []byte("Invalid Hash"))
		return
	}

	go func() {
		for {
			select {
			case a := <-offerClients[name].inbox:
				fmt.Println(a)
				messageUnwrapped := protocolUnwrapper(a)
				switch messageUnwrapped[0] {
				case "Offer":
					sendingMessage := protocolWrapper(messageUnwrapped[0], liMaker(messageUnwrapped[1]), messageUnwrapped[2])
					conn.WriteMessage(1, []byte(sendingMessage))
				case "Ice":

					fmt.Println("ICEICEICEICE")
					fmt.Println(a)
					conn.WriteMessage(1, []byte(a))
				}
			}
		}
	}()

	for {
		_, readBuffer, err = conn.ReadMessage()
		must(err)
		message = string(readBuffer)
		if message == "DONE" {
			return
		}

		fmt.Println(message)
		messageUnwrapped := protocolUnwrapper(message)

		switch messageUnwrapped[0] {
		case "Answer":
			offerClients[messageUnwrapped[1]].inbox <- messageUnwrapped[2]
		case "Ice":

			fmt.Println("ICEICEICEICE")
			fmt.Println(messageUnwrapped[2])
			sendTo := messageUnwrapped[1]
			sendingMessage := protocolWrapper(messageUnwrapped[0], name, messageUnwrapped[2])
			offerClients[sendTo].inbox <- sendingMessage

		}
	}
}

func videocallAnswerWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	must(err)
	var message string
	_, readBuffer, err := conn.ReadMessage()
	must(err)
	message = string(readBuffer)
	name, _, boolres := nameHashAnswerValidation(message)

	must(err)
	if !boolres {
		conn.WriteMessage(1, []byte("Invalid Hash"))
		return
	}

	// keep reading the offers
	lastUpdateLen := 0

	accu := ""
	go func() {

		for {
			if len(offerClients) != lastUpdateLen {
				lastUpdateLen = len(offerClients)
				for key := range offerClients {
					accu += liMaker(key)
				}
				sendingMessage := protocolWrapper("Offers", accu)
				conn.WriteMessage(1, []byte(sendingMessage))
				accu = ""
			}
		}
	}()
	go func() {

		for {

			_, readBuffer, err = conn.ReadMessage()
			must(err)
			message = string(readBuffer)
			if message == "DONE" {
				break
			}

			messageUnwrapped := protocolUnwrapper(message)

			switch messageUnwrapped[0] {
			case "Offer":
				sendTo := messageUnwrapped[1]
				sendingMessage := protocolWrapper(messageUnwrapped[0], name, messageUnwrapped[2])
				offerClients[sendTo].inbox <- sendingMessage
			case "Ice":
				fmt.Println("ICEICEICEICE")
				fmt.Println(messageUnwrapped[2])
				sendTo := messageUnwrapped[1]
				sendingMessage := protocolWrapper(messageUnwrapped[0], name, messageUnwrapped[2])
				offerClients[sendTo].inbox <- sendingMessage
			}
		}

	}()

	for {
		select {
		case a := <-answerClients[name].inbox:
			fmt.Println("sent to inbox")
			conn.WriteMessage(1, []byte(a))
		case <-time.After(60 * time.Second):
			fmt.Println("Timeout")
			return
			// create a channel to end the go routine above as well as this loop, also add the possibility for the one above to end it
			// end or be ended
		}
	}
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

// protocol: message type, then who knows it depends on the message type

func protocolUnwrapper(messageString string) []string {
	return strings.Split(messageString, protocolSep)
}

func protocolWrapper(componenets ...string) string {
	return strings.Join(componenets, protocolSep)
}
