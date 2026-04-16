package welcomeUi

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)

func ShowWelcomeUI() {
	// Load config
	cfg, _ := config.Load()

	mode := "Not configured"
	status := "Inactive"

	if cfg != nil {
		mode = fmt.Sprintf("%v", cfg.Intensity)
		if cfg.Enabled {
			status = "Active"
		}
	}

	// 🎨 Exact color palette
	mint := lipgloss.Color("#9FDBA1") // ✅ your exact color
	text := lipgloss.Color("#EDEDED")
	muted := lipgloss.Color("#eee") // subtle gray
	success := lipgloss.Color("#22C55E")

	// ===== Banner =====
	banner := `
   ██████╗  ██████╗  ██████╗ ██╗████████╗ ██████╗
  ██╔════╝ ██╔═══██╗██╔════╝ ██║╚══██╔══╝██╔═══██╗
  ██║      ██║   ██║██║  ███╗██║   ██║   ██║   ██║
  ██║      ██║   ██║██║   ██║██║   ██║   ██║   ██║
  ╚██████╗ ╚██████╔╝╚██████╔╝██║   ██║   ╚██████╔╝
   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝   ╚═╝    ╚═════╝
	`

	styledBanner := lipgloss.NewStyle().
		Foreground(mint).
		Bold(true).
		Render(banner)

	// Subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(muted).
		Render("Token Optimizer for AI CLI")

	// Status
	statusColor := muted
	if status == "Active" {
		statusColor = success
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render("● " + status)

	modeLine := lipgloss.NewStyle().
		Foreground(text).
		Render("Mode: " + mode)

	info := lipgloss.JoinVertical(
		lipgloss.Left,
		statusLine,
		modeLine,
	)

	// Commands
	arrow := lipgloss.NewStyle().
		Foreground(mint).
		Render("▸")

	cmd := func(label, desc string) string {
		return fmt.Sprintf(
			"%s %-18s %s",
			arrow,
			label,
			lipgloss.NewStyle().Foreground(muted).Render(desc),
		)
	}

	commands := lipgloss.JoinVertical(
		lipgloss.Left,
		cmd("cogito install", "setup hooks"),
		cmd("cogito config", "configure modes"),
		cmd("cogito uninstall", "remove hooks"),
		cmd("cogito --help", "show help"),
	)

	divider := lipgloss.NewStyle().
		Foreground(muted).
		Render("────────────────────────────")

	// Final Box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(muted).
		Padding(1, 3).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styledBanner,
				"",
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
