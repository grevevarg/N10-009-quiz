// Package model defines the quiz question types shared across the app.
// The JSON shapes here mirror multichoice.json and writtenlab.json exactly.
package model

// QuestionType identifies which of the three quiz question shapes a
// question uses.
type QuestionType string

const (
	TypeFillBlank   QuestionType = "fill_blank"
	TypeMultiChoice QuestionType = "multi_choice"
	TypeTable       QuestionType = "table"
)

// Chapter is a fixed textbook chapter, used to populate the chapter picker.
type Chapter struct {
	Number int
	Title  string
}

// Chapters lists all 21 chapters in book order.
var Chapters = []Chapter{
	{1, "Introduction to Networks"},
	{2, "The Open Systems Interconnection (OSI) Reference Model"},
	{3, "Networking Connectors and Wiring Standards"},
	{4, "The Current Ethernet Specifications"},
	{5, "Networking Devices"},
	{6, "Introduction to the Internet Protocol"},
	{7, "IP Addressing"},
	{8, "IP Subnetting, Troubleshooting IP, and Introduction to NAT"},
	{9, "Introduction to IP Routing"},
	{10, "Routing Protocols"},
	{11, "Switching and Virtual LANs"},
	{12, "Wireless Networking"},
	{13, "Remote Network Access"},
	{14, "Using Statistics and Sensors to Ensure Network Availability"},
	{15, "Organizational Documents and Policies"},
	{16, "High Availability and Disaster Recovery"},
	{17, "Data Center Architecture and Cloud Concepts"},
	{18, "Network Troubleshooting Methodology"},
	{19, "Network Software Tools and Commands"},
	{20, "Network Security Concepts"},
	{21, "Common Types of Attacks"},
}

// MultiChoiceQuestion is one entry from multichoice.json.
//
// Options is keyed by an internal, stable letter ("A".."F") that never
// changes for the lifetime of the data file. The UI is free to shuffle
// which of those internal keys is *displayed* as "A" in a given session;
// grading always happens against the internal key, never the display
// position.
type MultiChoiceQuestion struct {
	Chapter     int               `json:"chapter"`
	Type        QuestionType      `json:"type"`
	RandomOrder bool              `json:"random_order"`
	HasImages   bool              `json:"has_images"`
	Image       string            `json:"image,omitempty"`
	Prompt      string            `json:"prompt"`
	Options     map[string]string `json:"options"`
	Correct     []string          `json:"correct"`
}

// IsCorrect reports whether the given internal option key is an accepted
// answer. Per product decision, the user only ever picks one key even when
// len(Correct) > 1 (some questions have more than one acceptable answer).
func (q *MultiChoiceQuestion) IsCorrect(key string) bool {
	for _, c := range q.Correct {
		if c == key {
			return true
		}
	}
	return false
}

// OptionKeys returns the option keys in a stable, sorted order (A, B, C, ...).
// Callers that want a shuffled *display* order should permute the returned
// slice themselves rather than relying on map iteration order.
func (q *MultiChoiceQuestion) OptionKeys() []string {
	keys := make([]string, 0, len(q.Options))
	for _, l := range "ABCDEF" {
		k := string(l)
		if _, ok := q.Options[k]; ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// TableRow is one row of a TableQuestion. Cells holds the literal cell text,
// with "_" marking a blank; Answers maps the blank column's index (as a
// string, e.g. "1") to the exact-match expected value for that cell.
type TableRow struct {
	Cells   []string          `json:"cells"`
	Answers map[string]string `json:"answers"`
}

// WrittenLabQuestion is one entry from writtenlab.json. It covers both the
// "fill_blank" and "table" question types; only the fields relevant to
// Type are populated.
type WrittenLabQuestion struct {
	Chapter     int          `json:"chapter"`
	Type        QuestionType `json:"type"`
	RandomOrder bool         `json:"random_order"`
	HasImages   bool         `json:"has_images"`
	Image       string       `json:"image,omitempty"`
	Prompt      string       `json:"prompt"`

	// fill_blank only:
	Blanks []string `json:"blanks,omitempty"`

	// table only:
	Header []string   `json:"header,omitempty"`
	Rows   []TableRow `json:"rows,omitempty"`
}
