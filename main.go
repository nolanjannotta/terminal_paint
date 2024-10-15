package main

import (
	"fmt"
	"log"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	m := model{}
	m.initialize()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Paint")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeypress(msg)

	case tea.MouseMsg:
		m.handleMouseEvent(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.handleResize(msg)

	}

	return m, nil
}

func (m model) View() string {

	output := m.renderOutput()
	canvas := m.canvas.Render(output)

	colorPalette := m.renderColorPalette()

	tips := m.renderTips()

	selected := m.paint()

	options := lipgloss.JoinHorizontal(
		lipgloss.Center,
		fmt.Sprint(colorPalette, " ",
			tips,
			"⟬ selected: ",
			selected.Render(),
			" ⟭ ⟬ clear    ",
			m.params.erase.Render(),
			"    ",
			m.params.move.Render(),
			" ",
			m.offset.x,
			"x",
			m.offset.y,
			" ⟭ ⟬ save (todo) ⟭",
		))

	screen := lipgloss.JoinVertical(
		lipgloss.Left,
		canvas,
		options,
	)
	return screen
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *model) initialize() {
	for i := range 15 {
		m.colors = append(m.colors, i)
	}

	m.params = params{
		tip:   0,
		color: 1,
		move:  lipgloss.NewStyle().SetString("move"),
		erase: lipgloss.NewStyle().SetString("erase"),
	}

	m.offset = offset{x: 0, y: 0}
	m.pixelMap = make(map[[2]int]lipgloss.Style)

	m.tips = make([]tip, 6)

	m.tips = []tip{
		{char: "░", x: 51},
		{char: "▒", x: 54},
		{char: "▓", x: 57},
		{char: "■", x: 60},
		{char: "⬤", x: 63},
		{char: " ", x: 66},
		{char: ".", x: 69}, // nice
		{char: "◌", x: 72},
	}

}

func (m *model) renderColorPalette() string {
	var s string
	for _, color := range m.colors {
		s += lipgloss.NewStyle().
			Width(2).
			Height(1).
			Background(lipgloss.Color(strconv.Itoa(color))).Render()
	}
	return fmt.Sprint("⟬ colors: " + s + " ⟭")
}

func (m *model) renderTips() string {
	var s string
	for i, tip := range m.tips {
		if tip.char == " " {
			// background
			s += lipgloss.NewStyle().Width(1).Height(1).Background(lipgloss.Color(strconv.Itoa(m.params.color))).Render(m.tips[i].char) + "  "

		} else {
			s += lipgloss.NewStyle().Width(1).Height(1).Foreground(lipgloss.Color(strconv.Itoa(m.params.color))).Render(m.tips[i].char) + "  "
		}

	}
	return fmt.Sprint("⟬ tips: " + s + "⟭ ")
}

func (m *model) renderOutput() string {
	width := m.canvas.GetWidth()
	height := m.canvas.GetHeight()

	output := ""
	for y := 0; y < height; y++ {

		for x := 0; x < width; x++ {

			pixel := [2]int{x, y}
			pixel[0] += m.offset.x
			pixel[1] += m.offset.y
			if m.pixelMap[pixel].Value() == "" {
				output += m.pixelMap[pixel].Render(" ")
			} else {

				output += m.pixelMap[pixel].Render()
			}
		}
		if y < height-1 {
			output += "\n"
		}

	}
	return output
}
