package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

/* The server starts a server at port. The server creates a stdinReader goroutine. The server keeps a map of all open
 * connections. The values in the map contains channels to communicate with the writers.
 * The server is the controller, or master, of the running server program.
 */

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	writerCommandChan := make(chan writerCommand)
	writerInitChan <- writerInit{conn, writerCommandChan}
	writer(conn, writerCommandChan)

}

func server(port string, certFile string, keyFile string) {
	writerInitChan = make(chan writerInit)

	var httpServer http.Server
	httpServer.Addr = ":" + port
	httpServer.Handler = http.HandlerFunc(serverHandler)

	go func() {
		var err error
		if certFile == "" {
			fmt.Printf("The gowebsock server is listening on port %s\n", port)
			err = httpServer.ListenAndServe()
		} else {
			fmt.Printf("The gowebsock server is listening on port %s using TLS\n", port)
			err = httpServer.ListenAndServeTLS(certFile, keyFile)
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	stdinReaderChan := make(chan string)
	go stdinReader(stdinReaderChan)

	writerMap := make(map[string]writerInit)

	readerResultChan := make(chan readerResult)

	for {
		select {
		case writerInit := <-writerInitChan:
			conn := writerInit.conn
			go reader(conn, readerResultChan)

			address := conn.RemoteAddr().String()
			writerMap[address] = writerInit
			fmt.Printf("New connection: %s\n", address)
		case stdinMessage := <-stdinReaderChan:
			var messageType int
			data := ""
			switch stdinMessage {
			case "close":
				messageType = 8
			case "ping":
				messageType = 9
			case "pong":
				messageType = 10
			default:
				messageType = 1
				data = stdinMessage
			}
			for _, writerInit := range writerMap {
				writerInit.writerCommandChan <- writerCommand{false, messageType, []byte(data)}
			}
		case readerResult := <-readerResultChan:
			address := readerResult.conn.RemoteAddr().String()
			if readerResult.err == nil {
				output := address + ": type = " + messageTypeString(readerResult.messageType) + ", data = " + string(readerResult.data) + "\n"
				fmt.Printf(output)
			} else {
				writerInit := writerMap[address]
				writerInit.writerCommandChan <- writerCommand{true, 0, nil}
				delete(writerMap, address)
				fmt.Printf("%s: %s\n", address, readerResult.err)
			}
		}
	}
}
