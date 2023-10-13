package main

// :call jobstart("zellij run -c -- go run main.go")

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list         list.Model
	layoutItems  []layout
	sessionItems []session
	err          error
	msg          string
	bin          string
	args         []string
}

type LayoutsMsg struct {
	data []layout
}

type SessionsMsg struct {
	data []session
}

type errMsg struct {
	err error
}

type showMsg struct {
	msg string
}

type layout string
type session string

func (i layout) Title() string       { return string(i) }
func (i layout) Description() string { return fmt.Sprintf("New session with layout %v", i) }
func (i layout) FilterValue() string { return fmt.Sprintf("layout %v", i) }

func (i session) Title() string       { return string(i) }
func (i session) Description() string { return fmt.Sprintf("Attach to session %v", i) }
func (i session) FilterValue() string { return fmt.Sprintf("attach session %v", i) }

func reload() tea.Cmd {
	return tea.Batch(loadLayouts, fetchSessions)
}

func initialModel() model {
	items := []list.Item{}
	m := model{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = "Starting Zellij..."
	return m
}

func (m *model) buildItems() []list.Item {
	items := []list.Item{
		item{title: "New Default Session", desc: "New"},
	}
	if len(m.sessionItems) > 0 {
		for _, i := range m.sessionItems {
			items = append(items, i)
		}
	}
	if len(m.layoutItems) > 0 {
		for _, i := range m.layoutItems {
			items = append(items, i)
		}
	} else {
		i := item{title: "New Session With Layout", desc: "New"}
		items = append(items, i)
	}
	return items
}

// TODO: Load layouts
func loadLayouts() tea.Msg {
	// read the layouts directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = os.Getenv("HOME") + "/.config"
	}
	layoutDirName := xdgConfigHome + "/zellij/layouts"
	dir, err := os.ReadDir(layoutDirName)
	if err != nil {
		return errMsg{err: err}
	}
	// for each file, add it to the list
	list := []layout{}
	for _, file := range dir {
		name := file.Name()
		if strings.HasSuffix(name, ".kdl") {
			name = strings.TrimSuffix(name, ".kdl")
		}
		list = append(list, layout(name))
	}

	// return the list
	return LayoutsMsg{data: list}
}

func fetchSessions() tea.Msg {
	binaryName := "zellij"
	bin, err := exec.LookPath(binaryName)
	if err != nil {
		log.Fatalf("Error: %s binary not found in PATH", binaryName)
	}
	// Command to execute "zellij list-sessions" and capture its output
	cmd := exec.Command(bin, "list-sessions")

	// Set up a pipe to capture the command's output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		os.Exit(1)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		os.Exit(1)
	}

	// Read the output
	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Println("Error reading output:", err)
		os.Exit(1)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println("Command failed:", err)
		os.Exit(1)
	}

	// Convert the output to a string and split it into lines
	output := string(outputBytes)
	lines := strings.Split(output, "\n")

	// Create a slice to store the session names
	sessionNames := []session{}

	// Iterate through the lines and extract session names
	for _, line := range lines {
		if line != "" {
			// You may need to adjust the parsing logic depending on the zellij output format
			sessionNames = append(sessionNames, session(line))
		}
	}

	return SessionsMsg{data: sessionNames}
}

func (m model) Init() tea.Cmd {
	return reload()
}

func activateSelected(selected interface{}) (string, []string) {
	binaryName := "zellij"
	bin, err := exec.LookPath(binaryName)
	if err != nil {
		log.Fatalf("Error: %s binary not found in PATH", binaryName)
	}
	switch selected := selected.(type) {
	case item:
		return bin, []string{binaryName}
	case session:
		return bin, []string{binaryName, "a", string(selected)}
	case layout:
		// TODO: prompt for session name
		return bin, []string{binaryName, "--layout", string(selected)}
	}
	return "", nil
}

func (m model) exec() error {
	return syscall.Exec(m.bin, m.args, os.Environ())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case showMsg:
		m.msg = msg.msg

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case LayoutsMsg:
		m.layoutItems = msg.data
		return m, m.list.SetItems(m.buildItems())

	case SessionsMsg:
		m.sessionItems = msg.data
		return m, m.list.SetItems(m.buildItems())

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedItem()
			bin, args := activateSelected(selected)
			m.bin = bin
			m.args = args
			if args != nil {
				return m, tea.Quit
			}

		case "r":
			return m, reload()
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nAn error occurred: %v\n\n", m.err)
	}
	if m.msg != "" {
		return fmt.Sprintf("\nMessage: %v\n\n", m.msg)
	}
	return docStyle.Render(m.list.View())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if m, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	} else {
		fmt.Printf("And we're back! Here's the final model:\n\n%v\n", m)
		if m, ok := m.(model); ok {
			if m.args != nil {
				err = m.exec()

				if err != nil {
					log.Fatalf("Error: %v", err)
				}
			}
		}
	}
}
