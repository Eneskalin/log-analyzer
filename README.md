# Log Analyzer

[![Turkce](https://cdn-icons-png.flaticon.com/32/13980/13980321.png)](https://github.com/Eneskalin/log-analyzer/blob/main/READMETtr.md)
Turkish

A high-performance, real-time log analysis and monitoring tool developed in Go, utilizing the Bubble Tea (TUI) framework. This application is designed to parse complex log files, identify security threats, and provide live system monitoring.
## Technical Stack

- **Language:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
## Features

- **Real-Time Monitoring (Tailing):** Live tracking of multiple log files simultaneously with instant alert notifications for new entries.
- **Automated Log Analysis:** Generates comprehensive summaries including total line counts, event matching, and severity statistics (Critical, Error, Info).
- **Security Focused:** Designed to detect patterns like SQL Injection, XSS, and unauthorized access attempts based on customizable rules.
- **Export Capabilities:** Ability to export analysis results into CSV format for further investigation or reporting.





## Project Structure

- `/ui`: Contains the TUI logic, menu management, and view rendering.
- `/helpers`: Logic for log reading, pattern matching, and summary generation.
- `/models`: Internal data structures for logs, rules, and application state.
- `/config`: JSON configuration files for log paths and detection rules.


## Configuration & Customization

The system is highly flexible, allowing you to define which files to watch and what patterns to detect via JSON files in the `/config` directory.

You can map logical names to physical file paths. The application will resolve these paths relative to the configuration directory.

**Example:**
```json
{
  "logs": {
    "Web-Server": "logs/nginx/access.log",
    "System-Auth": "logs/auth.log",
    "Firewall": "logs/ufw.log"
  }
}
```




### Installation

 Clone the repository:
   ```bash
   git clone https://github.com/eneskalin/log-analyzer.git
   ```
   Build Project
   ```bash
   cd log-analyzer
   docker-compose up --build
   
  ```
  Open new Terminal
  ```bash
  docker-compose run log-analyzer
```


## Screen Shots


![Uygulama Ekran Görüntüsü](https://i.imgur.com/wRZP9lX.gif)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/8gEN0jJ.gif)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/7uDua0m.png)

![Uygulama Ekran Görüntüsü](https://i.imgur.com/PdnxTqv.gif)


