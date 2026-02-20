package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {

	return func (ps routing.PlayingState){
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerArmyMoves(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(m gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(m)
	}
}
