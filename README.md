
<div align="center">

```
  ██╗  ██╗██╗   ██╗██████╗ ██████╗ ██╗ █████╗      █████╗ ██╗
  ██║  ██║╚██╗ ██╔╝██╔══██╗██╔══██╗██║██╔══██╗    ██╔══██╗██║
  ███████║ ╚████╔╝ ██║  ██║██████╔╝██║███████║    ███████║██║
  ██╔══██║  ╚██╔╝  ██║  ██║██╔══██╗██║██╔══██║    ██╔══██║██║
  ██║  ██║   ██║   ██████╔╝██║  ██║██║██║  ██║    ██║  ██║██║
  ╚═╝  ╚═╝   ╚═╝   ╚═════╝ ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝   ╚═╝  ╚═╝╚═╝
```

**AI-Powered Attack Framework**

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go&logoColor=white)
![Gemini](https://img.shields.io/badge/Gemini-2.0_Flash-4285F4?style=flat-square&logo=google&logoColor=white)
![THC-Hydra](https://img.shields.io/badge/THC--Hydra-9.x-red?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=flat-square&logo=linux&logoColor=black)
![License](https://img.shields.io/badge/License-Educational_Only-orange?style=flat-square)

> Analyze a target image with **Gemini Vision API** → Generate a **smart wordlist** → Attack with **THC-Hydra**

</div>

---

## ⚠️ Legal Disclaimer

> **This tool is for educational purposes only.**
> Use it exclusively on **your own systems** or systems you have **explicit written permission** to test.
> Unauthorized access to computer systems is a criminal offense under applicable law.
> The developer assumes no legal responsibility for misuse.

---

## 🧠 How It Works

```
┌──────────────┐    ┌─────────────────────┐    ┌──────────────────┐    ┌──────────────┐
│  Upload      │───▶│  Gemini Vision API  │───▶│  Smart Wordlist  │───▶│  THC-Hydra   │
│  Image       │    │  OSINT Analysis     │    │  Generation      │    │  Attack      │
└──────────────┘    └─────────────────────┘    └──────────────────┘    └──────┬───────┘
                                                                               │
                                                        ┌──────────────────────▼───────┐
                                                        │  SQLite Tracker              │
                                                        │  NEVER retries a password    │
                                                        └──────────────────────────────┘
```

| Step | What Happens |
|------|-------------|
| **1. Input** | Provide an image, keyword text, or both — Gemini analyses each source |
| **2. Analysis** | Names, dates, pets, cities, hobbies, brands are extracted / expanded |
| **3. Wordlist** | Mutation rules produce thousands of prioritized password candidates |
| **4. Hydra Attack** | THC-Hydra runs in batches; stops the moment the password is found |
| **5. Tracking** | Every attempt is written to SQLite instantly — no password is ever tried twice |

---

## 🚀 Installation

### Requirements

- Linux (Ubuntu 20.04+, Debian, Kali, etc.)
- Go 1.22+
- THC-Hydra
- Gemini API Key ([get it free](https://aistudio.google.com))

### 1. Install THC-Hydra

```bash
sudo apt update && sudo apt install hydra -y
hydra -h   # Verify installation
```

### 2. Install Go Dependencies

```bash
go mod tidy
```

### 3. Set Your API Key

```bash
nano .env
```

Add this line:

```env
GEMINI_API_KEY=your_gemini_api_key_here
```

> 💡 Get a free key at [Google AI Studio](https://aistudio.google.com) → **Get API Key**

---

## 📖 Usage

### Basic Attack

```bash
go run main.go -i target.jpg -t 192.168.1.10 -s ssh -u admin
```

### All Flags

```
  -i, --image string      Path to target image file
      --text string       Keyword hints about the target (name, date, city, hobby, ...)
  -t, --target string     Target IP or domain
  -s, --service string    Protocol (ssh, ftp, rdp, ...)
  -u, --username string   Username
  -p, --port int          Port (optional)
      --api-key string    Gemini API key (reads from .env by default)
      --model string      Gemini model (default: gemini-2.0-flash)
      --session string    Session ID to resume
      --threads int       Parallel threads (default: 4)
      --batch-size int    Passwords per Hydra batch (default: 50)
      --dry-run           Generate wordlist only, do not run Hydra
```

### Input Modes

| Mode | Flag | Description |
|------|------|-------------|
| **Image only** | `-i target.jpg` | Gemini Vision analyzes the image visually |
| **Text only** | `--text "hints"` | Gemini expands your keywords into password candidates |
| **Combined** | `-i target.jpg --text "hints"` | Both analyses are merged — maximum coverage |

### Subcommands

```bash
# List all saved sessions
go run main.go sessions
```

### Examples

```bash
# SSH attack from image
go run main.go -i target.jpg -t 192.168.1.10 -s ssh -u root

# SSH attack from keywords only
go run main.go --text "john doe 1990 istanbul fenerbahce" -t 192.168.1.10 -s ssh -u john

# Combine image + extra keyword hints
go run main.go -i target.jpg --text "karabaş 1905" -t 192.168.1.10 -s ssh -u admin

# FTP on a custom port
go run main.go -i profile.png -t 10.0.0.5 -s ftp -u ftpuser -p 2121

# Generate wordlist only (no attack)
go run main.go --text "ali yilmaz 1992 trabzon" -t 192.168.1.10 -s ssh -u admin --dry-run

# List all sessions
go run main.go sessions

# Resume a paused attack
go run main.go --session sess_20260501_012710_abc123 -t 192.168.1.10 -s ssh -u admin

# High performance (16 threads)
go run main.go -i target.jpg -t 192.168.1.10 -s ssh -u admin --threads 16
```

---

## 🔐 Supported Protocols

| Protocol | Flag | Protocol | Flag |
|----------|------|----------|------|
| SSH | `ssh` | FTP | `ftp` |
| RDP | `rdp` | Telnet | `telnet` |
| SMTP | `smtp` | POP3 | `pop3` |
| IMAP | `imap` | MySQL | `mysql` |
| MSSQL | `mssql` | PostgreSQL | `postgres` |
| HTTP GET | `http-get` | HTTP POST Form | `http-post-form` |
| VNC | `vnc` | SMB | `smb` |

---

## 🧬 Wordlist Mutation Rules

| Rule | Input | Example Output |
|------|-------|----------------|
| Plain / Upper / Lower | `john` | `john`, `John`, `JOHN` |
| + Year | `john` + `1990` | `john1990`, `John1990` |
| + Number series | `john` | `john123`, `john1234` |
| + Special character | `john` | `john!`, `john@123` |
| Leet speak | `john` | `j0hn`, `j0h!n` |
| Combination | `john` + `doe` | `johndoe`, `john_doe` |
| Reversed | `john` | `nhoj`, `nhoj123` |
| Gemini suggestions | — | Added directly (highest priority) |

---

## 🗂️ Project Structure

```
hydria/
├── main.go                        ← Entry point (3 lines — loads .env, runs cmd)
├── go.mod / go.sum                ← Go module & dependencies
├── .env                           ← GEMINI_API_KEY (never commit this!)
├── .env.example                   ← Template
├── config.yaml                    ← Configuration
│
├── internal/
│   ├── config/
│   │   └── config.go              ← Config struct + Load()
│   │
│   ├── cmd/
│   │   ├── root.go                ← RootCmd, flag definitions, Execute()
│   │   ├── attack.go              ← Attack orchestration (analysis → wordlist → hydra)
│   │   └── sessions.go            ← `sessions` subcommand
│   │
│   ├── ui/
│   │   └── ui.go                  ← Terminal UI (lipgloss styling, tables, panels)
│   ├── vision/
│   │   └── vision.go              ← Gemini Vision API integration
│   ├── wordlist/
│   │   └── wordlist.go            ← Mutation engine + file I/O
│   ├── tracker/
│   │   └── tracker.go             ← SQLite attempt tracker (never retries)
│   ├── session/
│   │   └── session.go             ← Session CRUD operations
│   └── hydra/
│       └── runner.go              ← THC-Hydra subprocess wrapper
│
└── data/                          ← Auto-created at runtime
    ├── hydria.db                  ← SQLite database
    └── wordlists/                 ← Generated wordlist files
```

> **Design principle:** `main.go` stays at 3 lines of logic.
> Every responsibility lives in its own `internal/` package.

---

## ⚙️ Configuration (`config.yaml`)

```yaml
gemini:
  model: gemini-2.0-flash     # or gemini-1.5-pro for deeper analysis
  max_tokens: 8192

wordlist:
  max_size: 50000             # Maximum passwords to generate
  min_length: 4
  max_length: 20
  include_leet: true
  include_reverse: true
  include_combinations: true

hydra:
  threads: 4                  # Parallel threads
  timeout: 30                 # Connection timeout (seconds)
  batch_size: 50              # Passwords per Hydra invocation

session:
  auto_resume: true           # Offer to resume paused sessions automatically
```

---

## 💾 Session Management

Every attack is saved as a **session** and can be resumed at any point:

```bash
# List all sessions
./hydria sessions

# Resume where it stopped (Ctrl+C or crash)
./hydria --session sess_20260501_012710_abc123 -t 192.168.1.10 -s ssh -u admin
```

- Passwords tried in **previous sessions against the same target are also skipped**
- On Ctrl+C the session is marked `paused` — no progress is lost
- When the password is found the session is marked `completed`

---

## 🐛 Troubleshooting

| Error | Solution |
|-------|----------|
| `hydra: command not found` | Run `sudo apt install hydra` |
| `GEMINI_API_KEY not found` | Add your key to the `.env` file |
| `Unsupported image format` | Use `.jpg`, `.png`, or `.webp` |
| `Gemini returned invalid JSON` | Retry or try a different image |
| `Connection refused` | Check if the target service is running |

---

## 📦 Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/google/generative-ai-go` | Gemini Vision API client |
| `github.com/charmbracelet/lipgloss` | Terminal styling (colors, panels) |
| `github.com/schollz/progressbar/v3` | Progress bar |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/joho/godotenv` | `.env` file loader |
| `gopkg.in/yaml.v3` | `config.yaml` parser |
| `modernc.org/sqlite` | SQLite driver (pure Go, no CGO) |

---

<div align="center">

**HydrIA AI** — For ethical and authorized use only.

</div>
