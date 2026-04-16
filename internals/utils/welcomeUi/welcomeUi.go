package welcomeUi

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)

func ShowWelcomeUI() {
	// Load config safely
	cfg, _ := config.Load()

	mode := "Not configured"
	status := "Inactive"

	if cfg != nil {
		mode = fmt.Sprintf("%v", cfg.Intensity)
		if cfg.Enabled {
			status = "Active"
		}
	}

	// Colors
	var (
		accent   = lipgloss.Color("63")
		muted    = lipgloss.Color("240")
		highlight = lipgloss.Color("205")
		success  = lipgloss.Color("42")
		warn     = lipgloss.Color("214")
	)

	// Banner
	banner := `
   ██████╗  ██████╗  ██████╗ ██╗████████╗ ██████╗
  ██╔════╝ ██╔═══██╗██╔════╝ ██║╚══██╔══╝██╔═══██╗
  ██║      ██║   ██║██║  ███╗██║   ██║   ██║   ██║
  ██║      ██║   ██║██║   ██║██║   ██║   ██║   ██║
  ╚██████╗ ╚██████╔╝╚██████╔╝██║   ██║   ╚██████╔╝
   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝   ╚═╝    ╚═════╝
	`

	styledBanner := lipgloss.NewStyle().
		Foreground(highlight).
		Bold(true).
		Render(banner)

	subtitle := lipgloss.NewStyle().
		Foreground(muted).
		Render("Token Optimizer for AI CLI")

	// Status line
	statusColor := warn
	if status == "Active" {
		statusColor = success
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(fmt.Sprintf("● %s", status))

	modeLine := lipgloss.NewStyle().
		Foreground(accent).
		Render(fmt.Sprintf("Mode: %s", mode))

	info := lipgloss.JoinVertical(
		lipgloss.Left,
		statusLine,
		modeLine,
	)

	// Commands
	cmdStyle := lipgloss.NewStyle().Foreground(accent)

	commands := lipgloss.JoinVertical(
		lipgloss.Left,
		cmdStyle.Render("▸ cogito install   setup hooks"),
		cmdStyle.Render("▸ cogito config    configure modes"),
		cmdStyle.Render("▸ cogito uninstall remove hooks"),
		cmdStyle.Render("▸ cogito --help    show help"),
	)

	divider := lipgloss.NewStyle().
		Foreground(muted).
		Render("────────────────────────────")

	// Final box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 3).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styledBanner,
				subtitle,
				"",
				divider,
				info,
				"",
				commands,
			),
		)

	fmt.Println(box)
}
