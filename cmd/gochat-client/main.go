package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/gorilla/websocket"

	"github.com/jpparker/gochat/internal/pkg/service"
)

var (
	done = make(chan interface{})
	interrupt = make(chan os.Signal)
	text = make(chan string)
)

func main () {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	signal.Notify(interrupt, os.Interrupt)
	socketUrl := os.Getenv("SOCKET_URL")

	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatalln("Error connecting to GoChat server: ", err)
	}
	defer conn.Close()

	client := &service.Client {
		Conn: conn,
		Done: done,
		Text: text,
	}

	client.ReadUsernameInput()

	go client.ReceiveHandler()
	go client.ReadMessageInput()

	for {
		select {
		case msg := <- client.Text:
			err := client.SendMessage(msg)

			if err != nil {
				log.Println("Error sending message to WebSocket: ", err)
			}
		case <- interrupt:
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
					log.Println("Error during closing websocket: ", err)
					return
			}

			select {
			case <- client.Done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <- time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}