package main

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

func Life(b etk.Context) (etk.View, etk.React) {
	name := etk.State(b, "name", "game of life").Get()
	historyVar := etk.State(b, "history", []Board{blinker})
	history := historyVar.Get()
	stepVar := etk.State(b, "step", 0)
	step := stepVar.Get()

	return etk.Box(`
			[label=]
			board=`,
			etk.Text(ui.T(fmt.Sprintf("%s: %d / %d", name, step+1, len(history)))),
			etk.Text(ui.T(showBoard(history[step]))),
		), func(e term.Event) etk.Reaction {
			switch e {
			case term.K(ui.Left):
				if step > 0 {
					stepVar.Set(step - 1)
					return etk.Consumed
				}
			case term.K(ui.Right):
				if step == len(history)-1 {
					historyVar.Set(append(history, nextBoard(history[len(history)-1])))
				}
				stepVar.Set(step + 1)
				return etk.Consumed
			}
			return etk.Unused
		}
}

type Board [][]bool

func showBoard(b Board) string {
	var sb strings.Builder
	for i, row := range b {
		if i > 0 {
			sb.WriteByte('\n')
		}
		for _, cell := range row {
			if cell {
				sb.WriteRune('█')
			} else {
				sb.WriteRune(' ')
			}
		}
	}
	return sb.String()
}

func nextBoard(b Board) Board {
	newb := make(Board, len(b))
	for i := range newb {
		newb[i] = make([]bool, len(b[0]))
	}
	get := func(i, j int) int {
		if 0 <= i && i < len(b) && 0 <= j && j < len(b[0]) && b[i][j] {
			return 1
		}
		return 0
	}
	for i := range b {
		for j := range b[i] {
			liveNeighbors := get(i-1, j-1) + get(i-1, j) + get(i-1, j+1) + get(i, j-1) + get(i, j+1) + get(i+1, j-1) + get(i+1, j) + get(i+1, j+1)
			if b[i][j] {
				newb[i][j] = liveNeighbors == 2 || liveNeighbors == 3
			} else {
				newb[i][j] = liveNeighbors == 3
			}
		}
	}
	return newb
}

var blinker = Board{
	{false, false, false, false, false},
	{false, false, false, false, false},
	{false, true, true, true, false},
	{false, false, false, false, false},
	{false, false, false, false, false},
}

const truex = true

var pentadecathlon = Board{
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, truex, false, false, false, false, false},
	{false, false, false, false, truex, false, truex, false, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, truex, false, false, false, truex, false, false, false},
	{false, false, false, false, truex, false, truex, false, false, false, false},
	{false, false, false, false, false, truex, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
	{false, false, false, false, false, false, false, false, false, false, false},
}
