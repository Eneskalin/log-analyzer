package helpers

import (
	"encoding/csv"
	"fmt"
	"log-analyzer/models"
	"os"
	"strings"
	"time"
)

func GetSummary(filePath string) (models.LogSummary, error) {
	rules, err := ReadRules()
	if err != nil {
		return models.LogSummary{}, err
	}

	lines, err := ReadLogFile(filePath)
	if err != nil {
		return models.LogSummary{}, err
	}

	summary := models.LogSummary{
		FileName:      filePath,
		TotalLines:    len(lines),
		SeverityStats: make(map[string]int),
	}

	for _, line := range lines {
		for _, rule := range rules {
			if strings.Contains(line, rule.Match) {
				summary.MatchedEvents++
				summary.SeverityStats[rule.Severity]++

				detail := fmt.Sprintf("[%s] %s: %s", strings.ToUpper(rule.Severity), rule.Id, line)
				summary.Details = append(summary.Details, detail)
			}
		}
	}

	return summary, nil
}

func ExportToCSV(summary models.LogSummary) (string, error) {
    docsDir := "docs"
    if err := os.MkdirAll(docsDir, 0755); err != nil {
        return "", fmt.Errorf("klasör oluşturulamadı: %w", err)
    }

    timestamp := time.Now().Format("20060102_150405")
    
    // Windows path (D:\...) sorununu çözmek için sürücü harfini ayıklayalım
    cleanFileName := summary.FileName
    if len(cleanFileName) > 2 && cleanFileName[1] == ':' {
        cleanFileName = cleanFileName[3:] // D:\ kısmını atla
    }
    cleanFileName = strings.ReplaceAll(cleanFileName, "/", "_")
    cleanFileName = strings.ReplaceAll(cleanFileName, "\\", "_")

    fileName := fmt.Sprintf("%s/report_%s_%s.csv", docsDir, cleanFileName, timestamp)

    file, err := os.Create(fileName)
    if err != nil {
        return "", fmt.Errorf("dosya oluşturulamadı: %w", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    // Excel'in Türkçe karakterleri düzgün görmesi için (BOM ekleme)
    file.Write([]byte{0xEF, 0xBB, 0xBF}) 

    // Verileri yaz
    writer.Write([]string{"Dosya Adi", "Toplam Satir", "Tespit Edilen Olay"})
    writer.Write([]string{summary.FileName, fmt.Sprint(summary.TotalLines), fmt.Sprint(summary.MatchedEvents)})
    writer.Write([]string{""})
    writer.Write([]string{"Bulgu Detaylari"})

    for _, detail := range summary.Details {
        writer.Write([]string{detail})
    }

    writer.Flush()
    // Flush sonrası hata kontrolü kritik
    if err := writer.Error(); err != nil {
        return "", fmt.Errorf("csv yazma hatası: %w", err)
    }

    return fileName, nil
}
