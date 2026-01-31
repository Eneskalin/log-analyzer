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

		// Rastgele bir log dosyası seçimi (Simülasyon için)
		keys := make([]string, 0, len(pathConfig.Logs))
		for k := range pathConfig.Logs {
			keys = append(keys, k)
		}
		randomSource := keys[t.UnixNano()%int64(len(keys))]

		// Alert mesajı formatı
		alertMsg := fmt.Sprintf("\n> [!CAUTION]\n> ### 🚨 NEW ALERT FROM %s\n> **Time:** %s\n",
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
	m.fileList.Title = "Dosya Seçin"
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
	l.Title = "Log Dosyaları"
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

		// Read the log file and start tailing
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

func (m *Model) initializeFileOffsets() error {
	pathConfig, err := helpers.ReadPaths()
	if err != nil {
		return err
	}

	for _, logPath := range pathConfig.Logs {
		fileInfo, err := os.Stat(logPath)
		if err == nil {
			m.fileOffsets[logPath] = fileInfo.Size()
		}
	}
	return nil
}

func (m *Model) getNewLogContent() string {
	pathConfig, err := helpers.ReadPaths()
	if err != nil {
		return ""
	}

	var allContent strings.Builder
	allContent.WriteString("# 🚀 Canlı Sistem İzleme (Stream Mode)\n\n")
	allContent.WriteString(fmt.Sprintf("> **Son Güncelleme:** %s\n\n", time.Now().Format("15:04:05")))
	allContent.WriteString("---\n\n")

	hasNewContent := false

	for logName, logPath := range pathConfig.Logs {
		// Dosya bilgisini al
		fileInfo, err := os.Stat(logPath)
		if err != nil {
			continue
		}

		currentSize := fileInfo.Size()
		lastOffset := m.fileOffsets[logPath]

		// Eğer dosya büyüklüğü artmışsa yeni content var
		if currentSize > lastOffset {
			hasNewContent = true

			// Dosyayı oku
			content, err := os.ReadFile(logPath)
			if err != nil {
				continue
			}

			// Yeni kısımdan itibaren oku
			var newContent string
			if lastOffset > 0 {
				newContent = string(content[lastOffset:])
			} else {
				newContent = string(content)
			}

			// Offsenti güncelle
			m.fileOffsets[logPath] = currentSize

			// ALERT Başlığı - Yeni giriş göstergesi
			allContent.WriteString(fmt.Sprintf("## ⚠️ NEW ENTRY FROM %s\n", strings.ToUpper(logName)))

			// Satırları ayıkla
			lines := strings.Split(strings.TrimSpace(newContent), "\n")

			allContent.WriteString("```log\n")
			for _, line := range lines {
				if line != "" {
					allContent.WriteString(line + "\n")
				}
			}
			allContent.WriteString("```\n\n")
		}
	}

	if !hasNewContent {
		// Yeni content yoksa durum mesajı göster
		allContent.WriteString("> ⏳ Yeni log girişi bekleniyor...\n")
	}

	return allContent.String()
}

func (m *Model) streamCmd() tea.Cmd {
	pathConfig, err := helpers.ReadPaths()
	if err != nil {
		return func() tea.Msg { return err }
	}

	// Optimize: Initialize file offsets once
	if len(m.fileOffsets) == 0 {
		m.initializeFileOffsets()
	}

	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		newAlerts := make([]string, 0, 10)

		for name, path := range pathConfig.Logs {
			// Optimize: Use os.Stat instead of opening file first
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			currSize := info.Size()
			lastSize, exists := m.fileOffsets[path]

			// Skip if file hasn't changed
			if !exists || currSize <= lastSize {
				if !exists {
					m.fileOffsets[path] = currSize
				}
				continue
			}

			// Optimize: Only open file when we know there's new content
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			// Read only new content
			newContent := content[lastSize:]
			if len(newContent) == 0 {
				continue
			}

			// Update offset
			m.fileOffsets[path] = currSize

			// Process lines
			lines := strings.Split(strings.TrimSpace(string(newContent)), "\n")
			lineCount := 0

			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && lineCount < 5 { // Limit alerts per file per tick
					alert := fmt.Sprintf("**[%s]** 🚨 `%s`\n> %s",
						t.Format("15:04:05"),
						strings.ToUpper(name),
						trimmed)
					newAlerts = append(newAlerts, alert)
					lineCount++
				}
			}
		}

		if len(newAlerts) > 0 {
			return tailMsg(strings.Join(newAlerts, "\n\n"))
		}
		return tailMsg("")
	})
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
