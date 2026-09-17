stack:
Go
Bubble Tea (+ bubbles for lists/checkboxes/tables, lipgloss for styling) for TUI
go:embed for bundling quiz JSON + images into the binary — no external data files to ship or lose
single static binary per platform, cross-compiled via GOOS/GOARCH from one machine

platforms:
full cross-platform (linux, macOS, windows)
image questions require a terminal that supports an inline image protocol (iTerm2 / Kitty / Sixel).
these are complex network topology diagrams — not ASCII-reducible — so this is a hard requirement, not a nice-to-have.
  - macOS: iTerm2 supports both its own inline protocol and sixel natively. no extra setup needed.
  - windows: PowerShell is a shell, not a terminal emulator — image support depends on what it's running IN.
    - Windows Terminal (v1.22+) supports sixel. REQUIRED.
    - legacy console host (conhost.exe / old blue PowerShell window) supports nothing. must be called out explicitly in install docs/readme: "requires Windows Terminal," link to install, since a beginner-targeted tool can't assume the user picked their terminal deliberately.
  - linux: kitty, wezterm, foot etc. work. plain vte-based terminals (gnome-terminal, etc.) do not — same capability-detection story applies.
  - fallback when no protocol detected: don't attempt any ascii/sixel degradation, just tell the user the terminal doesn't support images and point them at the image file path directly (still available on disk, unpacked from the binary into a temp/cache dir at runtime if needed).

image loading:
module: github.com/BourgeoisBear/rasterm — encodes to iTerm2 / Kitty / Sixel, and ships capability detection: IsItermCapable(), IsKittyCapable(), IsSixelCapable(), IsTmuxScreen()
- detect once at startup, cache in app state, don't re-probe per question
- IsTmuxScreen() matters: tmux/screen commonly break or need passthrough config for these protocols even on an otherwise-capable terminal — detect it and warn rather than silently failing
- images embedded via go:embed alongside the quiz JSON, referenced by the same relative path convention as the source data (./images/filename.png)

typo checker proposal (Damerau-Levenshtein, not plain Levenshtein):
plain Levenshtein scores a transposed pair of letters (e.g. "recieve" vs "receive") as distance 2, not 1 — and a swapped pair is probably the single most common real typo. Damerau-Levenshtein treats a transposition as one edit, so it actually catches the fat-finger case we care about. cost of implementing it ourselves over plain Levenshtein is trivial.

func CheckAnswer(userInput, target string) string {
    u := strings.ToLower(strings.TrimSpace(userInput))
    t := strings.ToLower(strings.TrimSpace(target))

    if u == t {
        return "EXACT"
    }

    dist := damerauLevenshtein([]rune(u), []rune(t)) // operate on runes, not bytes — avoids UTF-8 miscounts on accented chars
    if len([]rune(t)) > 6 && dist <= 1 {
        return "TYPO_WARN" // warn, don't dock score
    }

    return "INCORRECT"
}

notes:
- len(t) > 6 guard kept from the original proposal — it's doing more work than it looks like. below that length, distance-1 is too large a proportional change and risks accepting a different-but-similar short answer (e.g. two 3-letter acronyms one edit apart) as a "typo."
- multi-word fill-in-blank answers: normalize whitespace before distance calc so a stray double-space isn't burning the distance-1 budget that should be reserved for an actual character typo.
- todo: eyeball the finished answer bank once built for any pair of distinct valid answers that are themselves distance 1 apart (false-positive risk), since this fails silently otherwise.

general:
UI motivation: TUI tool to teach people to not be scared of the terminal. you will need to get comfortable in it, so get comfortable now.

because some users are beginners, automatic install scripts should be reviewed. when adding command to PATH the user should be able to launch it as the name "quiztia"

let user choose chapters to review. build a list and randomize it. randomized lists of 1 element always have the same order, no need for if statement. performance cost negligible.

chapter picker should be presented as columnar list, checkbox + color highlighting on selected chapters. sort order asc/desc should be swappable. at index 0 place "select all chapters". there are 21 chapters in total.

build quiz of chapter questions in separate lists based on question type. randomize each list — unless a question block's source data says otherwise (see `random_order` below, which is authored per chapter/section, not assumed globally true). present question, wait for user input, increment score if correct, present textbook answer, wait for user to progress, present next question.

present score at end of run, perhaps even overview of their answers + textbook answers formatted as a scrollable table.

source data format (as authored, before import):
each chapter/question block carries flags as comment-style lines, e.g.:
  ## random_order = true
  ## is_table = true
  ## has_images = true
these get parsed into the JSON fields below by an import step. todo: write that importer — it needs to strip nano's line-number gutter and reflow the wrapped lines back into single logical lines/rows before parsing.

question types:

"fill in the blank"
complete a provided sentence using 1-n user-inputted strings. case insensitive.
typo tolerance per the Damerau-Levenshtein checker above — key purpose is making sure acronyms/initialisms stick, only warn at fat-fingering, never dock score for it.

"multi-choice"
user gets presented with multiple choices and has to pick one — or more than one. some answer-key questions accept multiple correct answers, so the schema needs an array of correct indices, and the UI needs to switch input mode accordingly (single-select "press a letter" vs multi-select "toggle checkboxes, confirm with enter") and should tell the user up front which mode it's in ("select all that apply").
options prepended with letters A-D (or further, if a multi-answer question has more than 4 options — confirm max option count once the full answer bank is in).
option order shuffles between sessions to mitigate position memorization.

"table"
some labs are answered as a table rather than a sentence — a header row plus data rows where either column can hold the blank (`_`) depending on the row; it's not always "answer goes in the right column." example: OSI-layer/device table has blanks in column 2 throughout, but the CIDR/subnet-mask table alternates which column is blank per row. schema needs to record, per row, which cell(s) are blanks and what the correct answer for each is — not just "this row has a blank."
row order respects the same `random_order` flag as the other types.

"has_images"
question (of any of the above types) additionally references a figure the user needs to answer against. rendered via the image-loading pipeline above before the question prompt. image filename stored relative to ./images/, matching the source data convention.

json schema (draft, one object per question):

{
  "chapter": 5,
  "type": "fill_blank" | "multi_choice" | "table",
  "random_order": true,
  "has_images": false,
  "image": "6.png",              // only present if has_images is true, relative to ./images/
  "prompt": "string",

  // fill_blank only:
  "blanks": ["answer1", "answer2"],   // 1-n, in order

  // multi_choice only:
  "choices": ["choice text", "..."],
  "correct": [0, 2],                  // indices into choices; len > 1 means multi-select

  // table only:
  "header": ["Description", "Device or OSI layer"],
  "rows": [
    {
      "cells": ["This device sends and receives information about the Network layer.", "_"],
      "answers": { "1": "Network layer" }   // keyed by blank column index
    },
    {
      "cells": ["_", "255.0.0.0"],
      "answers": { "0": "/8" }
    }
  ]
}

todo:
- importer/parser: source (nano-pasted, flag-commented) -> JSON above
- confirm max option count for multi-choice (A-D assumed, may need to extend)
- confirm answer-key format for table cells that themselves accept typo tolerance (e.g. "/8" vs "8" for CIDR) — probably exact-match only for table cells, typo tolerance for fill_blank/multi_choice free text
