package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "23234"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	pty, _, _ := s.Pty()
	renderer := bubbletea.MakeRenderer(s)

	m := model{}
	m.term = pty
	m.renderer = renderer

	m.initialize()

	return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseAllMotion()}
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
			" ⟭",
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
		move:  m.renderer.NewStyle().SetString("move"),
		erase: m.renderer.NewStyle().SetString("erase"),
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
		s += m.renderer.NewStyle().
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
			s += m.renderer.NewStyle().Width(1).Height(1).Background(lipgloss.Color("7")).Render() + "  "

		} else {
			s += m.tips[i].char + "  "
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

			pixel := [2]int{x + m.offset.x, y + m.offset.y}
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
