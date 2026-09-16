stack:
python + textual (or if textual looks like ass, use curses)

typo checker proposal:
import difflib

```py
def check_answer(user_input: str, target: str) -> str:
    u, t = user_input.strip().lower(), target.strip().lower()
    
    if u == t:
        return "EXACT"
    
    # Calculate similarity ratio
    ratio = difflib.SequenceMatcher(None, u, t).ratio()
    
    # Require minimum length (> 6 chars) and a high similarity threshold (~85%+)
    if len(t) > 6 and ratio >= 0.85:
        return "TYPO_WARN"
        
    return "INCORRECT"
```

general:
UI motivation: TUI tool to teach people to not be scared of the terminal. You will need to get comfortable in it, so get comfortable now.

let user choose chapters to review. build a list and randomize it. randomized lists of 1 element always have the same order, no need for if statement. performance cost neglible.

chapter picker should be presented as columnar list, checkbox + color highlighting on selected chapters. sort order asc/desc should be swappable. at index 0 place "select all chapters". there are 21 chapters in total.

build quiz of chapter questions in two separate lists based on question type. randomize each list. present question, wait for user input, increment score if correct, present textbook answer, wait for user to progress, present next question.

present score at end of run, perhaps even overview of their answers + textbook answers formatted as a scrollable table.

question types:

todo: json schema

"fill in the blank"
Complete a provided sentence using 1-n user inputted strings. case insensitive?
Accept typos with levehenstein distance 1 with a warning if string is over 6 characters? not interested in heavy NLP solutions. key purpose: make sure that acronyms and initialisms stick, only warn at fat-fingering but don't dock the score.

"multi-choice"
user gets presented with multiple choices and has to pick one.
options should be prepended with letters A-D.
answers should shift positions between sessions to mitigate bruteforce position memorization.
