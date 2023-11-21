package gol

var ExecuteTurns = "GolOp.ExecuteTurns"
var ExecuteTimer = "GolOp.Timer"
var ExecuteKey = "GolOp.KeyOp"
var KillProcess = "GolOp.Kill"
var ResumeProcess = "GolOp.Resume"

type Request struct {
	World [][]uint8
	P     Params
}

type Response struct {
	NewWorld    [][]uint8
	Final       FinalTurnComplete
	CurrentTurn int
}

type KeyPress struct {
	Key rune
	P   Params
}
type ReportAlive struct {
	Alive AliveCellsCount
}
