package gol

import (
	"fmt"
	"net/rpc"
	"os"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
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

	for i := 0; i < p.ImageHeight; i++ {
		for k := 0; k < p.ImageWidth; k++ {
			if world[i][k] == 255 {
				c.events <- CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: k, Y: i}}
			}
		}
	}

	//stopChan := make(chan bool, 1)
	//server := "54.164.24.79:8030"
	server := "127.0.0.1:8030"
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	sdlTicker := time.NewTicker(30 * time.Millisecond)
	defer sdlTicker.Stop()
	//var mutex sync.Mutex
	//var wg sync.WaitGroup
	//
	//go func() {
	//	for {
	//		select {
	//		case <-sdlTicker.C:
	//			//mutex.Lock()
	//			wg.Add(1)
	//			//mutex.Unlock()
	//			req := new(Request)
	//			req.P = p
	//			res := new(Response)
	//			client.Call("GolOp.Live", req, res)
	//			//res.NewWorld = make([][]uint8, req.P.ImageHeight)
	//			//for i := range res.NewWorld {
	//			//	res.NewWorld[i] = make([]uint8, req.P.ImageWidth)
	//			//}
	//			for i := 0; i < p.ImageHeight; i++ {
	//				for k := 0; k < p.ImageWidth; k++ {
	//					if res.NewWorld[i][k] != world[i][k] {
	//						c.events <- CellFlipped{CompletedTurns: res.CurrentTurn, Cell: util.Cell{X: k, Y: i}}
	//					}
	//				}
	//			}
	//			c.events <- TurnComplete{CompletedTurns: res.CurrentTurn}
	//			copyWhole(world, res.NewWorld)
	//			wg.Done()
	//		}
	//	}
	//}()
	go func() {
		for {
			select {
			case <-ticker.C:
				request := new(Request)
				reportAlive := new(ReportAlive)
				client.Call(ExecuteTimer, request, reportAlive)
				c.events <- reportAlive.Alive
			case k := <-c.keyP:
				switch k {
				case 's':
					fmt.Println("s is pressed (save)")
					key := KeyPress{Key: 's', P: p}
					res := new(Response)
					client.Call(ExecuteKey, key, res)
					c.ioCommand <- ioCommand(0)
					fileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, res.CurrentTurn)
					c.ioFilename <- fileName
					go func() {
						for i := 0; i < p.ImageHeight; i++ {
							for k := 0; k < p.ImageWidth; k++ {
								c.ioOutput <- res.NewWorld[i][k]
							}
						}
						c.events <- ImageOutputComplete{CompletedTurns: res.CurrentTurn, Filename: fileName}
					}()

				case 'k':
					fmt.Println("k is pressed (kill)")
					key := KeyPress{Key: 'k', P: p}
					res := new(Response)
					client.Call(ExecuteKey, key, res)
					c.ioCommand <- ioCommand(0)
					fileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, res.CurrentTurn)
					c.ioFilename <- fileName
					go func() {
						for i := 0; i < p.ImageHeight; i++ {
							for k := 0; k < p.ImageWidth; k++ {
								c.ioOutput <- res.NewWorld[i][k]
							}
						}
						c.events <- ImageOutputComplete{CompletedTurns: res.CurrentTurn, Filename: fileName}
					}()
					client.Call(KillProcess, key, res)
					c.events <- res.Final
					// Make sure that the Io has finished any output before exiting.
					c.ioCommand <- ioCheckIdle
					<-c.ioIdle
					c.events <- StateChange{res.CurrentTurn, Quitting}
					// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
					close(c.events)
					os.Exit(0)
				case 'p':
					fmt.Println("p is pressed (pause)")
					key := KeyPress{Key: 'p', P: p}
					res := new(Response)
					client.Call(ExecuteKey, key, res)
					fmt.Println("done")
					c.ioCommand <- ioCommand(0)
					fileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, res.CurrentTurn)
					fmt.Println(fileName)
					c.ioFilename <- fileName
					go func() {
						for i := 0; i < p.ImageHeight; i++ {
							for k := 0; k < p.ImageWidth; k++ {
								c.ioOutput <- res.NewWorld[i][k]
							}
						}
						c.events <- ImageOutputComplete{CompletedTurns: res.CurrentTurn, Filename: fileName}
						c.events <- StateChange{CompletedTurns: res.CurrentTurn, NewState: Paused}
					}()
				OUTER:
					for {
						select {
						case k := <-c.keyP:
							if k == 'p' {
								client.Call(ResumeProcess, key, res)
								fmt.Println("Continuing")
								c.events <- StateChange{CompletedTurns: res.CurrentTurn, NewState: Executing}
								break OUTER
							}
						}
					}

				case 'q':
					fmt.Println("q is pressed (quit)")
					key := KeyPress{Key: 'q', P: p}
					res := new(Response)
					client.Call(ExecuteKey, key, res)
					//// Make sure that the Io has finished any output before exiting.
					//c.ioCommand <- ioCheckIdle
					//<-c.ioIdle
					//c.events <- StateChange{res.CurrentTurn, Quitting}
					//// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
					//close(c.events)
					os.Exit(0)
				}
				//case <-stopChan:
				//	return
			}
		}
	}()
	makeCall(client, world, p, c)
	turn := p.Turns
	// TODO: Report the final state using FinalTurnCompleteEvent.
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	//mutex.Lock()
	//wg.Wait()
	//mutex.Unlock()
	close(c.events)
	//stopChan <- true
}

func makeCall(client *rpc.Client, world [][]uint8, p Params, c distributorChannels) {
	request := Request{world, p}
	response := new(Response)
	client.Call(ExecuteTurns, request, response)
	c.events <- response.Final
	c.ioCommand <- ioCommand(0)
	fileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, p.Turns)
	c.ioFilename <- fileName
	for i := 0; i < p.ImageHeight; i++ {
		for k := 0; k < p.ImageWidth; k++ {
			c.ioOutput <- response.NewWorld[i][k]
		}
	}
	c.events <- ImageOutputComplete{CompletedTurns: p.Turns, Filename: fileName}

}
func copyWhole(dst, src [][]uint8) {
	for i := range src {
		copy(dst[i], src[i])
	}
}
