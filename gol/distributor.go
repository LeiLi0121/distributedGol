package gol

import (
	"fmt"
	"flag"
	"net/rpc"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyP       <-chan rune
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	fileName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioCommand <- ioCommand(1)
	c.ioFilename <- fileName
	//world recieve the image from io-----------------------------------------------------------------------------------
	world := make([][]uint8, p.ImageHeight)
	for i := range world {
		world[i] = make([]uint8, p.ImageWidth)
	}

	for i := 0; i < p.ImageHeight; i++ {
		for k := 0; k < p.ImageWidth; k++ {
			world[i][k] = <-c.ioInput
		}
	}
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()
	makeCall(client,world,p,c)
	turn:=p.Turns
	// TODO: Report the final state using FinalTurnCompleteEvent.
	//c.ioCommand <- ioCommand(0)
	//fileName = fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
	//c.ioFilename <- fileName
	//for i := 0; i < p.ImageHeight; i++ {
	//	for k := 0; k < p.ImageWidth; k++ {
	//		c.ioOutput <- world[i][k]
	//	}
	//}
	
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func makeCall(client *rpc.Client, world [][]uint8,p Params, c distributorChannels) {
	
	request := Request{world,p}
	response := new(Response)
	client.Call(ExecuteTurns, request, response)
	c.events<-response.Final
	
}

