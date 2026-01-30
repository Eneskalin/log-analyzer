package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

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
	Quitting      bool
	fileOffsets   map[string]int64
	RecentAlerts  []string
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
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case analysisResultMsg:
		m.AnalysisData = string(msg)
		m.viewport.SetContent(m.AnalysisData)
		return m, nil

	case tailMsg:
		if m.currentScreen == streamScreen {
			newContent := string(msg)
			if newContent != "" {
				m.RecentAlerts = append(m.RecentAlerts, newContent)
				if len(m.RecentAlerts) > 20 {
					m.RecentAlerts = m.RecentAlerts[1:]
				}
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# 🚀 CANLI SİSTEM TAKİP\n\n**Son Güncelleme:** %s\n\n", time.Now().Format("15:04:05")))

			sb.WriteString("## 🚨 ALERT'LER\n\n")
			if len(m.RecentAlerts) > 0 {
				for i := len(m.RecentAlerts) - 1; i >= 0; i-- {
					sb.WriteString(m.RecentAlerts[i] + "\n\n")
				}
			} else {
				sb.WriteString("> 🟢 **Normal Durum** - Yeni hareketlilik bekleniyor...\n")
			}

			m.AnalysisData = renderWithGlamour(sb.String(), m.viewport.Width)
			m.viewport.SetContent(m.AnalysisData)

			return m, m.streamCmd()
		}
		return m, nil

	case error:
		m.AnalysisData = "Hata oluştu: " + msg.Error()
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
					if choice == "Çıkış" {
						m.Quitting = true
						return m, tea.Quit
					}
					if choice == "Gerçek Zamanlı İzleme (Tailing)" {
						m.currentScreen = streamScreen
						return m, m.streamCmd()
					}
					m.currentScreen = fileListScreen
					m.LoadLogFiles()
					return m, nil
				}
			} else if m.currentScreen == fileListScreen {
				i, ok := m.fileList.SelectedItem().(item)
				if ok {
					m.SelectedFile = string(i)
					if m.Choice == "CSV Raporu Oluştur" {
						return m, m.exportCSVCmd()
					}
					m.currentScreen = analysisScreen
					return m, m.runAnalysisCmd()
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
		return quitTextStyle.Render("Uygulamadan çıkılıyor...")
	}

	switch m.currentScreen {
	case mainMenuScreen:
		return "\n" + m.mainMenu.View()
	case fileListScreen:
		return "\n" + m.fileList.View() + "\n\n   ESC: Geri Dön"
	case analysisScreen, streamScreen:
		return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerView()
	}
	return ""
}

func InitialModel() *Model {
	items := []list.Item{
		item("Dosya Bazlı Analiz Özetleri"),
		item("Gerçek Zamanlı İzleme (Tailing)"),
		item("CSV Raporu Oluştur"),
		item("Çıkış"),
	}
	l := list.New(items, itemDelegate{}, 30, listHeight)
	l.Title = "Log Analyzer - Ana Menü"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	emptyFileList := list.New([]list.Item{}, itemDelegate{}, 30, listHeight)
	emptyFileList.Title = "Dosya Seçin"
	emptyFileList.SetShowStatusBar(false)
	emptyFileList.SetFilteringEnabled(false)

	return &Model{
		currentScreen: mainMenuScreen,
		mainMenu:      l,
		fileList:      emptyFileList,
		fileOffsets:   make(map[string]int64),
	}
}
