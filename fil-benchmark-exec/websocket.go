package fil_benchmark_exec

import (
	"fil-benchmark/datastructures"
	"fil-benchmark/external-testbed"
	"github.com/gorilla/websocket"
	"log"
)

func connect2LotusWsApi(node datastructures.Node, interrupt chan bool, input chan string, output chan string) {

	// Get Api Token
	token := external_testbed.GetTestBed().GetLotusApiToken(node)

	// Define URL
	u := "ws://" + node.Ip + ":3000/rpc/v0?token=" + token

	log.Printf("connecting to %s", u)

	// Connect to websocket
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Forward response chan
	go func(o chan string) {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			output <- string(message)
		}
	}(output)

	// Forward request chan
	for {
		select {
		case <-done:
			return
		case t := <-input:
			err := c.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case b := <-interrupt:
			if b {
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				return
			}
		}
	}
}
