package util

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
)

func JsonMarshalWhatever(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func ToTeaCmd[T tea.Msg](f func() T) tea.Cmd {
	return func() tea.Msg {
		return f()
	}
}
