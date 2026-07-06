package ui

import (
	"fmt"
	"io"
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

type cardDelegate struct{ focused bool }

func (d cardDelegate) Height() int                             { return 1 }
func (d cardDelegate) Spacing() int                            { return 0 }
func (d cardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	card, ok := item.(cardItem)
	if !ok {
		return
	}
	s := "  " + card.title
	if d.focused && index == m.Index() {
		s = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> " + card.title)
	}
	fmt.Fprint(w, s)
}

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
		// list style
		l := list.New(items, cardDelegate{focused: i == 0}, 0, 0)
		l.Title = col.Title
		l.SetShowFilter(false)
		l.SetShowHelp(false)

		cols[i] = l
	}
	return model{
		columns:     cols,
		frontmatter: f.Frontmatter,
		config:      f.Config,
		filePath:    filePath,
	}
}

func (m *model) syncDelegates() {
	for i := range m.columns {
		m.columns[i].SetDelegate(cardDelegate{focused: i == m.focusedCol})
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
			m.columns[i].SetSize(colW, msg.Height-4)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			if m.focusedCol > 0 {
				m.focusedCol--
				m.syncDelegates()
			}
			return m, nil
		case "right", "l":
			if m.focusedCol < len(m.columns)-1 {
				m.focusedCol++
				m.syncDelegates()
			}
		case "L":
			if m.focusedCol < len(m.columns)-1 {
				cursor := m.columns[m.focusedCol].Cursor()
				toMove := m.columns[m.focusedCol].SelectedItem()
				m.columns[m.focusedCol].RemoveItem(cursor)

				// clamp in case last item was removed
				if n := len(m.columns[m.focusedCol].Items()); n > 0 {
					m.columns[m.focusedCol].Select(min(cursor, n-1))
				}

				dest := m.columns[m.focusedCol+1].Cursor()
				m.columns[m.focusedCol+1].InsertItem(dest, toMove)
				m.focusedCol++
				m.columns[m.focusedCol].Select(dest)
				m.syncDelegates()
			}
			return m, nil
		case "H":
			if m.focusedCol > 0 {
				cursor := m.columns[m.focusedCol].Cursor()
				toMove := m.columns[m.focusedCol].SelectedItem()
				m.columns[m.focusedCol].RemoveItem(cursor)

				// clamp in case last item was removed
				if n := len(m.columns[m.focusedCol].Items()); n > 0 {
					m.columns[m.focusedCol].Select(min(cursor, n-1))
				}

				dest := m.columns[m.focusedCol-1].Cursor()
				m.columns[m.focusedCol-1].InsertItem(dest, toMove)
				m.focusedCol--
				m.columns[m.focusedCol].Select(dest)
				m.syncDelegates()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.columns[m.focusedCol], cmd = m.columns[m.focusedCol].Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	focused := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 3, 1, 1)
	unfocused := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 3, 1, 1)
	cols := make([]string, len(m.columns))
	for i, col := range m.columns {
		s := col.View()
		if m.focusedCol == i {
			s = focused.Render(s)
		} else {
			s = unfocused.Render(s)
		}
		cols[i] = s
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
