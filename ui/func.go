package ui

import (
	"fmt"
	"log-analyzer/helpers"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) tailAllLogsCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*800, func(t time.Time) tea.Msg {
		pathConfig, _ := helpers.ReadPaths()

		keys := make([]string, 0, len(pathConfig.Logs))
		for k := range pathConfig.Logs {
			keys = append(keys, k)
		}
		randomSource := keys[t.UnixNano()%int64(len(keys))]

		alertMsg := fmt.Sprintf("\n> [!CAUTION]\n> ###  NEW ALERT FROM %s\n> **Time:** %s\n",
			strings.ToUpper(randomSource),
			t.Format("15:04:05"))

		return tailMsg(alertMsg)
	})
}

func (m *Model) LoadLogFiles() error {
	pathConfig, err := helpers.ReadPaths()
	if err != nil {
		return err
	}
	var items []list.Item
	for name := range pathConfig.Logs {
		items = append(items, item(name))
	}
	m.fileList = list.New(items, itemDelegate{}, 30, listHeight)
	m.fileList.Title = "Select File"
	m.fileList.SetShowStatusBar(false)
	m.fileList.SetFilteringEnabled(false)
	return nil
}

func loadLogFiles(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files
}

func createFileList(files []string) list.Model {
	items := []list.Item{}
	for _, file := range files {
		items = append(items, item(file))
	}

	const defaultWidth = 30
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Log Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	return l
}

func (m Model) runAnalysisCmd() tea.Cmd {
	return func() tea.Msg {
		pathConfig, _ := helpers.ReadPaths()
		selectedPath := pathConfig.Logs[m.SelectedFile]

		summary, err := helpers.GetSummary(selectedPath)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Analysis Summary: %s\n\n", summary.FileName))
		sb.WriteString(fmt.Sprintf("* **Total Lines:** %d\n", summary.TotalLines))
		sb.WriteString(fmt.Sprintf("* **Total Detection:** %d\n\n", summary.MatchedEvents))

		sb.WriteString("### Severity Stats\n")
		if len(summary.SeverityStats) == 0 {
			sb.WriteString("* No incident detected.\n")
		} else {
			for sev, count := range summary.SeverityStats {
				sb.WriteString(fmt.Sprintf("* **%s**: %d\n", strings.ToUpper(sev), count))
			}
		}
		sb.WriteString("\n")

		sb.WriteString("### Details\n")
		for _, detail := range summary.Details {
			sb.WriteString("- " + detail + "\n")
		}

		return analysisResultMsg(renderWithGlamour(sb.String(), m.viewport.Width))
	}
}

func (m Model) tailLogCmd() tea.Cmd {
	return func() tea.Msg {
		pathConfig, err := helpers.ReadPaths()
		if err != nil {
			return err
		}
		selectedPath := pathConfig.Logs[m.SelectedFile]

		content, err := os.ReadFile(selectedPath)
		if err != nil {
			return err
		}

		return analysisResultMsg(string(content))
	}
}

func (m Model) liveAnalysis() tea.Cmd {
	return func() tea.Msg {
		return tailMsg("Live analysis started")
	}
}





func WaitForLog(sub chan tailMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func StartLogWorker(sub chan tailMsg) {
	// OPTİMİZASYON: Fileoffsets ve config'i worker'da tut
	fileOffsets := make(map[string]int64)

	// Config'i yükle
	pathConfig, err := helpers.ReadPaths()
	if err != nil {
		sub <- tailMsg("Error loading paths: " + err.Error())
		return
	}

	// Başlangıç offsetlerini al
	for _, logPath := range pathConfig.Logs {
		if info, err := os.Stat(logPath); err == nil {
			fileOffsets[logPath] = info.Size()
		}
	}

	// OPTİMİZASYON: İlk okumada fazla CPU kullanımını önle
	ticker := time.NewTicker(500 * time.Millisecond) // 500ms daha responsive
	defer ticker.Stop()

	const bufferSize = 65536 // 64KB buffer - daha az syscall

	for range ticker.C {
		newAlerts := make([]string, 0, 10)

		for name, path := range pathConfig.Logs {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			currSize := info.Size()
			lastSize, exists := fileOffsets[path]

			// Değişiklik yoksa geç
			if exists && currSize == lastSize {
				continue
			}

			// Dosya küçüldüyse (Rotate) offset sıfırla
			var seekOffset int64 = 0
			if exists && currSize >= lastSize {
				seekOffset = lastSize
			}

			// Dosyayı aç ve oku (64KB buffer)
			file, err := os.Open(path)
			if err != nil {
				continue
			}

			_, err = file.Seek(seekOffset, 0)
			if err != nil {
				file.Close()
				continue
			}

			buf := make([]byte, bufferSize)
			n, err := file.Read(buf)
			file.Close()

			// Offseti hemen güncelle
			fileOffsets[path] = currSize

			if n == 0 {
				continue
			}

			newContent := string(buf[:n])
			trimmedContent := strings.TrimSpace(newContent)

			if trimmedContent == "" {
				continue
			}

			// OPTİMİZASYON: Manuel satır parsing (Split yerine)
			lineCount := 0
			start := 0
			timeStr := time.Now().Format("15:04:05")
			nameLower := strings.ToUpper(name)

			for i := 0; i < len(trimmedContent) && lineCount < 5; i++ {
				if trimmedContent[i] == '\n' {
					line := strings.TrimSpace(trimmedContent[start:i])
					if line != "" {
						alert := fmt.Sprintf("**[%s]** 🚨 `%s`\n> %s",
							timeStr, nameLower, line)
						newAlerts = append(newAlerts, alert)
						lineCount++
					}
					start = i + 1
				}
			}
			if start < len(trimmedContent) && lineCount < 5 {
				line := strings.TrimSpace(trimmedContent[start:])
				if line != "" {
					alert := fmt.Sprintf("**[%s]** 🚨 `%s`\n> %s",
						timeStr, nameLower, line)
					newAlerts = append(newAlerts, alert)
				}
			}
		}

		// Eğer yeni log varsa kanala gönder
		if len(newAlerts) > 0 {
			select {
			case sub <- tailMsg(strings.Join(newAlerts, "\n\n")):
			default:
				// Kanal dolu ise, mesajı atla (back pressure)
			}
		}
	}
}



func (m *Model) headerView() string {
	title := titleStyle.Render("Log Summary: " + m.SelectedFile)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m Model) exportCSVCmd() tea.Cmd {
	return func() tea.Msg {
		pathConfig, _ := helpers.ReadPaths()
		selectedPath := pathConfig.Logs[m.SelectedFile]

		summary, err := helpers.GetSummary(selectedPath)
		if err != nil {
			return err
		}

		fileName, err := helpers.ExportToCSV(summary)
		if err != nil {
			return err
		}

		result := fmt.Sprintf("# ✅ Rapor Hazırlandı\n\nAnaliz başarıyla tamamlandı ve CSV dosyası oluşturuldu.\n\n**Dosya Adı:** `%s`", fileName)
		return analysisResultMsg(renderWithGlamour(result, m.viewport.Width))
	}
}
