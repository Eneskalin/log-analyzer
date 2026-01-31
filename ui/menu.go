package ui

import (
	"fmt"
	"io"
	"strings"
	"time"
	"log-analyzer/models"
	"log-analyzer/helpers"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type screen int
type tailMsg string
type analysisResultMsg string

const (
	mainMenuScreen screen = iota
	fileListScreen
	analysisScreen
	streamScreen
)

const listHeight = 14

var (
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
	}
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("170")).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	infoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(str))
}

type Model struct {
	currentScreen screen
	mainMenu      list.Model
	fileList      list.Model
	viewport      viewport.Model
	Choice        string
	SelectedFile  string
	AnalysisData  string
	spinner       spinner.Model
	isLoading     bool
	loadingMsg    string
	Quitting      bool
	RecentAlerts  []string	
	sub           chan tailMsg 
	PathConfig    models.PathConfig 
}

func renderWithGlamour(content string, width int) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)
	out, _ := r.Render(content)
	return out
}

func (m *Model) Init() tea.Cmd {
	m.sub = make(chan tailMsg)

	if config, err := helpers.ReadPaths(); err == nil {
		m.PathConfig = config
	}

	go StartLogWorker(m.sub)

	return tea.Batch(
		m.spinner.Tick,
		WaitForLog(m.sub), 
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
case spinner.TickMsg:
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd 
	case analysisResultMsg:
		m.isLoading = false
		m.AnalysisData = string(msg)
		m.viewport.SetContent(m.AnalysisData)
		return m, nil

case tailMsg:
		if m.currentScreen == streamScreen {
			newContent := string(msg)
			if newContent != "" {
				m.RecentAlerts = append(m.RecentAlerts, newContent)
				if len(m.RecentAlerts) > 50 {
					m.RecentAlerts = m.RecentAlerts[1:]
				}
				var sb strings.Builder
				header := fmt.Sprintf("\n 🚀 LIVE SYSTEM MONITORING - Last Update: %s\n", time.Now().Format("15:04:05"))
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render(header))
				sb.WriteString(strings.Repeat("─", m.viewport.Width) + "\n\n")

				for i := len(m.RecentAlerts) - 1; i >= 0; i-- {
					sb.WriteString(m.RecentAlerts[i] + "\n\n")
				}
				
				m.AnalysisData = sb.String()
				m.viewport.SetContent(m.AnalysisData)
			}
		}
		return m, WaitForLog(m.sub)
	case error:
		m.isLoading = false
		m.AnalysisData = "Error: " + msg.Error()
		m.viewport.SetContent(m.AnalysisData)
		return m, nil

	case tea.WindowSizeMsg:
		m.mainMenu.SetWidth(msg.Width)
		if m.fileList.Items() != nil {
			m.fileList.SetWidth(msg.Width)
		}
		m.viewport = viewport.New(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "esc":
			if m.currentScreen == analysisScreen {
				m.currentScreen = fileListScreen
				m.AnalysisData = ""
				return m, nil
			}
			if m.currentScreen == fileListScreen || m.currentScreen == streamScreen {
				m.currentScreen = mainMenuScreen
				return m, nil
			}
			m.Quitting = true
			return m, tea.Quit

		case "enter":
			if m.currentScreen == mainMenuScreen {
				i, ok := m.mainMenu.SelectedItem().(item)
				if ok {
					choice := string(i)
					m.Choice = choice
					if choice == "Exit" {
						m.Quitting = true
						return m, tea.Quit
					}
					if choice == "Monitoring (Tailing)" {
						m.currentScreen = streamScreen
						m.AnalysisData = ""
						m.RecentAlerts = []string{}
						m.viewport.SetContent(lipgloss.NewStyle().
                    Foreground(lipgloss.Color("240")).
                    Render("\n   Connecting, waiting for new logs...")) 
						return m, nil
					}
					m.currentScreen = fileListScreen
					m.LoadLogFiles()
					return m, nil
				}
			} else if m.currentScreen == fileListScreen {
				i, ok := m.fileList.SelectedItem().(item)
				if ok {
					m.isLoading = true
					m.SelectedFile = string(i)
					m.currentScreen = analysisScreen
					randomIndex := time.Now().UnixNano() % int64(len(spinners))
            		m.resetSpinner(spinners[randomIndex])
					if m.Choice == "Save CSV" {
						m.loadingMsg = "CSV file in progress..."
						return m, tea.Batch(m.exportCSVCmd(), m.spinner.Tick)
					}
					m.loadingMsg = "Summary is in progress..."
					return m, tea.Batch(m.runAnalysisCmd(), m.spinner.Tick)
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.currentScreen == mainMenuScreen {
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	} else if m.currentScreen == fileListScreen {
		m.fileList, cmd = m.fileList.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m *Model) View() string {
	if m.Quitting {
		return quitTextStyle.Render("Exiting the application...")
	}
	if m.isLoading {
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.loadingMsg)
	}

	switch m.currentScreen {
	case mainMenuScreen:
		return "\n" + m.mainMenu.View()
	case fileListScreen:
		return "\n" + m.fileList.View() + "\n\n   ESC: Back"
	case analysisScreen, streamScreen:
		return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerView()
	}
	return ""
}

func InitialModel() *Model {
	s := spinner.New()
	s.Spinner = spinner.Pulse
	items := []list.Item{
		item("Log Summary"),
		item("Monitoring (Tailing)"),
		item("Save CSV"),
		item("Exit"),
	}
	l := list.New(items, itemDelegate{}, 30, listHeight)
	l.Title = "-- Log Analyzer --"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	emptyFileList := list.New([]list.Item{}, itemDelegate{}, 30, listHeight)
	emptyFileList.Title = "Select File"
	emptyFileList.SetShowStatusBar(false)
	emptyFileList.SetFilteringEnabled(false)

	return &Model{
		spinner:       s,
		isLoading:     false,
		currentScreen: mainMenuScreen,
		mainMenu:      l,
		fileList:      emptyFileList,
	}
}

func (m *Model) resetSpinner(s spinner.Spinner) {
    m.spinner = spinner.New()
    m.spinner.Style = spinnerStyle 
    m.spinner.Spinner = s
}
