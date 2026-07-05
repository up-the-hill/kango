package ui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/up-the-hill/kango/internal/parser"
)

type cardItem struct {
	done  bool
	title string
}

func (c cardItem) FilterValue() string { return c.title }
func (c cardItem) Title() string       { return c.title }
func (c cardItem) Description() string { return "" }

type model struct {
	columns     []list.Model
	focusedCol  int
	frontmatter []string
	config      []string
	filePath    string
}

func newModel(f parser.File, filePath string) model {
	cols := make([]list.Model, len(f.Main))
	for i, col := range f.Main {
		items := make([]list.Item, len(col.Cards))
		for j, card := range col.Cards {
			items[j] = cardItem{done: card.Done, title: card.Title}
		}
		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = col.Title
		cols[i] = l
	}
	return model{
		columns:     cols,
		frontmatter: f.Frontmatter,
		config:      f.Config,
		filePath:    filePath,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		colW := msg.Width / len(m.columns)
		for i := range m.columns {
			m.columns[i].SetSize(colW, msg.Height)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			if m.focusedCol > 0 {
				m.focusedCol--
			}
			return m, nil
		case "right", "l":
			if m.focusedCol < len(m.columns)-1 {
				m.focusedCol++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.columns[m.focusedCol], cmd = m.columns[m.focusedCol].Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	cols := make([]string, len(m.columns))
	for i, col := range m.columns {
		cols[i] = col.View()
	}
	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	v.AltScreen = true
	return v
}

func Main(f parser.File, filePath string) {
	p := tea.NewProgram(newModel(f, filePath))
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
