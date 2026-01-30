package main

import (
	"fmt"
	"log-analyzer/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mainModel := ui.InitialModel()
	program := tea.NewProgram(mainModel)
	_, err := program.Run()
	if err != nil {
		fmt.Printf("Hata: %v\n", err)
		return
	}

}
