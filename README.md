# N10-009-quiz

Terminal quiz app for studying the CompTIA Network+ (N10-009) exam. 

Runs in your terminal.

## 1. Get a terminal

Most modern terminals works. If yours renders TUIs badly, use one of these:

**Windows** — Windows Terminal
```
winget install Microsoft.WindowsTerminal
```

**macOS** — iTerm2 (or Ghostty)
```
brew install --cask iterm2
```

**Linux** — Ghostty (or use your existing terminal as long as it supports Kitty Graphics Protocol)
Install from: https://ghostty.org/download

## 2. Download and run

Grab the binary for your OS from the latest release:

https://github.com/grevevarg/N10-009-quiz/releases/latest

- `quiztia-windows-amd64.exe` — Windows
- `quiztia-macos-arm64` — macOS (Apple Silicon)
- `quiztia-linux-amd64` — Linux

Run it from your terminal:

```
./quiztia-linux-amd64      # or the file you downloaded
```

(macOS/Linux: `chmod +x quiztia-macos-arm64` or `chmod +x quiztia-linux-amd64` the file first if it won't execute.)

That's it. Follow the on-screen prompts.

## Optional: clone and run from source

Need Go 1.27+ first: https://go.dev/dl/

```
git clone https://github.com/grevevarg/N10-009-quiz.git
cd N10-009-quiz
go run .
```
