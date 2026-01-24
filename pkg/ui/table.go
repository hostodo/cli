package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hostodo/hostodo-cli/pkg/api"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ListMode ViewMode = iota
	DetailMode
)

// TableModel represents the Bubble Tea model for the instances table
type TableModel struct {
	table            table.Model
	instances        []api.Instance
	selectedInstance int
	mode             ViewMode
	quitting         bool
}

// NewTableModel creates a new table model with instances
func NewTableModel(instances []api.Instance) TableModel {
	columns := []table.Column{
		{Title: "ID", Width: 12},
		{Title: "Hostname", Width: 25},
		{Title: "IP Address", Width: 16},
		{Title: "Status", Width: 14},
		{Title: "Power", Width: 12},
		{Title: "RAM", Width: 10},
		{Title: "CPU", Width: 6},
		{Title: "Disk", Width: 8},
	}

	rows := make([]table.Row, len(instances))
	for i, instance := range instances {
		rows[i] = table.Row{
			truncate(instance.InstanceID, 12),
			truncate(instance.Hostname, 25),
			instance.MainIP,
			instance.Status,
			instance.PowerStatus,
			fmt.Sprintf("%d MB", instance.RAM),
			fmt.Sprintf("%d", instance.VCPU),
			fmt.Sprintf("%d GB", instance.Disk),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(instances)+2, 20)),
	)

	// Custom styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(highlightColor).
		Bold(true)

	t.SetStyles(s)

	return TableModel{
		table:     t,
		instances: instances,
		mode:      ListMode,
	}
}

// Init initializes the table model
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			if m.mode == DetailMode {
				// Go back to list view
				m.mode = ListMode
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.mode == ListMode {
				// Switch to detail view
				m.selectedInstance = m.table.Cursor()
				if m.selectedInstance < len(m.instances) {
					m.mode = DetailMode
					return m, nil
				}
			} else {
				// Return to list view
				m.mode = ListMode
				return m, nil
			}
		}
	}

	// Only update table when in list mode
	if m.mode == ListMode {
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// View renders the table or detail view
func (m TableModel) View() string {
	if m.quitting {
		return ""
	}

	if m.mode == DetailMode {
		// Show detail view
		var sb strings.Builder
		sb.WriteString(FormatInstanceDetail(&m.instances[m.selectedInstance]))
		sb.WriteString("\n")
		sb.WriteString(HelpStyle.Render("Press Enter to return to list • q/Esc to quit"))
		sb.WriteString("\n")
		return sb.String()
	}

	// Show list view
	var sb strings.Builder

	// Title
	title := TitleStyle.Render("Hostodo Instances")
	sb.WriteString(title + "\n\n")

	// Table
	sb.WriteString(m.table.View() + "\n\n")

	// Help text
	help := HelpStyle.Render("↑/↓: Navigate • Enter: Details • q: Quit")
	sb.WriteString(help + "\n")

	return sb.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
