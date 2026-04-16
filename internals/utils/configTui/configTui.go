package configTui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)

type Step int

const (
	StepMode Step = iota
	StepDone
)

type Model struct {
	step   Step
	cursor int

	modes []string

	selectedMode string
}

func InitialModel() Model {
	return Model{
		step:  StepMode,
		modes: []string{"Lite", "Normal", "Ultra"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			if m.cursor < len(m.modes)-1 {
				m.cursor++
			}

		case "enter":
			m.selectedMode = m.modes[m.cursor]
			m.step = StepDone
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.step == StepDone {
		return fmt.Sprintf(
			"\n✅ Config Saved:\nMode: %s\n\n",
			m.selectedMode,
		)
	}

	s := "Select Mode:\n\n"

	for i, choice := range m.modes {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nUse ↑ ↓ and Enter\n"
	return s
}

func mapMode(mode string) config.Intensity {
	switch mode {
	case "Lite":
		return config.IntensityLite
	case "Ultra":
		return config.IntensityUltra
	default:
		return config.IntensityNormal
	}
}

// Run this inside your Install() or init command
func RunConfigTUI() {
	p := tea.NewProgram(InitialModel())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	final := finalModel.(Model)

	cfg := &config.Config{
		Enabled:   true,
		Intensity: mapMode(final.selectedMode),
	}

	if err := config.Save(cfg); err != nil {
		fmt.Println("❌ Failed to save config:", err)
	} else {
	}
}
