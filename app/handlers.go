package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	m.width, m.height = msg.Width, msg.Height
	m.canvas = lipgloss.NewStyle().Width(msg.Width - 2).Height(msg.Height - 3).BorderStyle(lipgloss.RoundedBorder())

}

func (m *model) handleMouseEvent(msg tea.MouseMsg) {
	switch msg.Action {
	case 0: // mouse press down

		// start drawing
		if msg.X > 0 && msg.Y > 0 && msg.X <= m.canvas.GetWidth() && msg.Y <= m.canvas.GetHeight() {
			m.isDrawing = true
			pixel := [2]int{msg.X + m.offset.x - 1, msg.Y + m.offset.y - 1}

			if m.params.erase.GetUnderline() {
				m.pixelMap[pixel] = lipgloss.NewStyle()
				return
			}

			if m.params.move.GetUnderline() {
				m.offset.startingX, m.offset.startingY = msg.X, msg.Y
				return

			}
			if (m.pixelMap[pixel].GetBackground() != lipgloss.NoColor{} && (m.tips[m.params.tip].char == " ")) {
				m.pixelMap[pixel] = m.paint()
				return
			}

			if (m.pixelMap[pixel].GetBackground() != lipgloss.NoColor{}) {
				m.pixelMap[pixel] = m.overlay(pixel)
				return
			}
			m.pixelMap[pixel] = m.paint()
			return
		}

		if msg.Y == m.height-1 {
			// select colors
			if isBetweenOrEqual(msg.X, 39, 10) {
				minInput := 10
				maxInput := 40
				minOutput := 0 // pointless but for clarity
				maxOutput := 15
				m.params.color = (msg.X-minInput)*(maxOutput-minOutput)/(maxInput-minInput) + minOutput
				m.params.move = m.params.move.Underline(false).Bold(false)
				m.params.erase = m.params.erase.Underline(false).Bold(false)
				return

			}

			// clear canvas
			if isBetweenOrEqual(msg.X, 95, 99) {
				m.pixelMap = make(map[[2]int]lipgloss.Style)
				m.offset.x, m.offset.y = 0, 0
				m.params.move = m.params.move.Underline(false).Bold(false)
				m.params.erase = m.params.erase.Underline(false).Bold(false)
				return

			}
			// erase
			if isBetweenOrEqual(msg.X, 104, 108) {
				state := m.params.erase.GetUnderline()
				m.params.erase = m.params.erase.Underline(!state).Bold(!state)
				m.params.move = m.params.move.Underline(false).Bold(false)
				return

			}
			// move
			if isBetweenOrEqual(msg.X, 113, 116) {
				state := m.params.move.GetUnderline()
				m.params.move = m.params.move.Underline(!state).Bold(!state)
				m.params.erase = m.params.erase.Underline(false).Bold(false)
				return

			}

			// this is ugly, but since the x,y offset coordinates can change on the screen, which will push the 'save' button to the right
			// we need to make sure the x,y coordinates of 'save' button are lined up

			xstr, ystr := strconv.Itoa(m.offset.x), strconv.Itoa(m.offset.y)
			save := 126 + len(xstr) + len(ystr) - 2
			// save
			if isBetweenOrEqual(msg.X, save, save+3) {
				// this doesnt actually save an image, only saves the output string in a text file
				saveImage(m.renderOutput())
				return

			}

			// if none of those if statements pass, we select a tip
			for i, tip := range m.tips {
				if msg.X == tip.x {
					m.params.tip = i
				}
			}
			m.params.move = m.params.move.Underline(false).Bold(false)
			m.params.erase = m.params.erase.Underline(false).Bold(false)

		}

	case 1: // mouse press up
		m.isDrawing = false
		return

	case 2: // mouse moving
		// handle drawing
		if m.isDrawing {
			pixel := [2]int{msg.X + m.offset.x - 1, msg.Y + m.offset.y - 1}
			if m.params.erase.GetUnderline() {
				m.pixelMap[pixel] = lipgloss.NewStyle()
				return
			}

			if m.params.move.GetUnderline() {
				deltaX := msg.X - m.offset.startingX
				deltaY := msg.Y - m.offset.startingY
				m.offset.x, m.offset.y = m.offset.x-int(deltaX), m.offset.y-int(deltaY)
				m.offset.startingX, m.offset.startingY = msg.X, msg.Y
				return

			}
			if (m.pixelMap[pixel].GetBackground() != lipgloss.NoColor{} && (m.tips[m.params.tip].char == " ")) {
				m.pixelMap[pixel] = m.paint()
				return
			}

			if (m.pixelMap[pixel].GetBackground() != lipgloss.NoColor{}) {
				m.pixelMap[pixel] = m.overlay(pixel)
				return
			}

			m.pixelMap[pixel] = m.paint()
			return
		}

	}

}

func (m *model) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) paint() lipgloss.Style {
	if m.tips[m.params.tip].char == " " {
		return lipgloss.NewStyle().Background(lipgloss.Color(strconv.Itoa(m.params.color))).SetString(" ")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(m.params.color))).SetString(m.tips[m.params.tip].char)

}

func (m *model) overlay(pixel [2]int) lipgloss.Style {
	return m.pixelMap[pixel].SetString(m.tips[m.params.tip].char).Foreground(lipgloss.Color(strconv.Itoa(m.params.color)))

}

func isBetweenOrEqual(a, b, c int) bool {
	// returns true if a is equal to or between b and c
	return float64(a) <= math.Max(float64(b), float64(c)) && float64(a) >= math.Min(float64(b), float64(c))
}

func saveImage(img string) bool {
	files, err := os.ReadDir("images")
	if err != nil {
		if err := os.Mkdir("images", 0755); err != nil {
			return false
		}
		files, _ = os.ReadDir("images")

	}
	newFile, err := os.Create(fmt.Sprint("images/", len(files)+1, ".txt"))

	if err != nil {
		return false
	}
	defer newFile.Close()

	data := []byte(img)
	_, err = newFile.Write(data)
	return err == nil
}
