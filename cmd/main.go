package main

import (
	"fmt"
	"limiu82214/lazyAppleMusic/internal/model"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var dump *os.File
	if _, ok := os.LookupEnv("DEBUG"); ok {
		var err error
		dump, err = os.OpenFile("tmp/debug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(1)
		}
	}

	//p := tea.NewProgram(internal.InitialModel(dump))
	p := tea.NewProgram(model.InitialTopModel(dump))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
