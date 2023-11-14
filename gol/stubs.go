package gol

var ExecuteTurns= "GolOp.ExecuteTurns"

type Request struct {
	World [][]uint8
	P Params
}

type Response struct {
	NewWorld [][]uint8
	Final FinalTurnComplete
}

