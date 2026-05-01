// internal/wordlist/wordlist.go
// Smart wordlist generator from Gemini analysis results.
// Applies mutation rules and orders passwords by likelihood.
package wordlist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hydria-ai/hydria/internal/vision"
)

// Common suffixes ordered by frequency
var commonSuffixes = []string{
	"123", "1234", "12345", "!", "@", "#",
	"1", "2", "0", "00", "11", "99",
	"123!", "!123", "@123", "1!", "2!", "0!",
	"321", "2023", "2024", "2025", "2026",
	".", "..", "_", "__",
}

var commonPrefixes = []string{
	"123", "!", "@", "#", "the", "my", "new", "old",
}

var leetMap = map[rune][]string{
	'a': {"@", "4"},
	'e': {"3"},
	'i': {"1", "!"},
	'o': {"0"},
	's': {"$", "5"},
	't': {"7"},
	'l': {"1"},
	'g': {"9"},
}

var specialChars = []string{"!", "@", "#", "$", "*", ".", "_", "-"}

func capitalizeVariants(word string) []string {
	lower := strings.ToLower(word)
	upper := strings.ToUpper(word)
	title := strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
	seen := map[string]struct{}{word: {}, lower: {}, upper: {}, title: {}}
	out := []string{word, lower, upper, title}
	_ = seen
	return out
}

func leetVariants(word string) []string {
	lower := strings.ToLower(word)
	runes := []rune(lower)
	var out []string
	for i, ch := range runes {
		if replacements, ok := leetMap[ch]; ok {
			for _, r := range replacements {
				leet := string(runes[:i]) + r + string(runes[i+1:])
				out = append(out, leet)
				if len(leet) > 0 {
					out = append(out, strings.ToUpper(leet[:1])+leet[1:])
				}
			}
		}
	}
	return out
}

func withSuffixes(base string, dates []string) []string {
	var out []string
	cap1 := strings.ToUpper(base[:1]) + strings.ToLower(base[1:])
	for _, s := range commonSuffixes {
		out = append(out, base+s, cap1+s)
	}
	for _, d := range dates {
		out = append(out, base+d, cap1+d, d+base)
	}
	return out
}

func withPrefixes(base string) []string {
	var out []string
	for _, p := range commonPrefixes {
		out = append(out, p+base)
	}
	return out
}

func combinations(words []string) []string {
	top := words
	if len(top) > 5 {
		top = top[:5]
	}
	separators := []string{"", "_", "."}
	var out []string
	for i, w1 := range top {
		for j, w2 := range top {
			if i == j {
				continue
			}
			for _, sep := range separators {
				combo := w1 + sep + w2
				out = append(out, combo)
				if len(combo) > 0 {
					out = append(out, strings.ToUpper(combo[:1])+combo[1:])
				}
			}
		}
	}
	return out
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Options controls wordlist generation behaviour.
type Options struct {
	MinLength          int
	MaxLength          int
	MaxSize            int
	IncludeLeet        bool
	IncludeReverse     bool
	IncludeCombinations bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		MinLength:           4,
		MaxLength:           20,
		MaxSize:             50000,
		IncludeLeet:         true,
		IncludeReverse:      true,
		IncludeCombinations: true,
	}
}

// Generate produces a prioritized password list from Gemini analysis data.
func Generate(analysis vision.AnalysisResult, opts Options) []string {
	seen := make(map[string]struct{})
	var ordered []string

	add := func(word string) {
		w := strings.TrimSpace(word)
		if w == "" {
			return
		}
		if len([]rune(w)) < opts.MinLength || len([]rune(w)) > opts.MaxLength {
			return
		}
		if _, exists := seen[w]; exists {
			return
		}
		seen[w] = struct{}{}
		ordered = append(ordered, w)
	}

	addAll := func(words []string) {
		for _, w := range words {
			add(w)
		}
	}

	dates := analysis.Dates
	names := analysis.Names

	// 1. Gemini's direct suggestions — highest priority
	for _, hint := range analysis.CustomHints {
		add(hint)
		addAll(capitalizeVariants(hint))
	}

	// 2. Names
	for _, name := range names {
		addAll(capitalizeVariants(name))
		addAll(withSuffixes(name, dates))
	}

	// 3. Other categories
	other := append(analysis.Pets, analysis.Locations...)
	other = append(other, analysis.Interests...)
	other = append(other, analysis.Brands...)
	other = append(other, analysis.Numbers...)

	for _, word := range other {
		addAll(capitalizeVariants(word))
		addAll(withSuffixes(word, dates))
	}

	// 4. Dates alone
	for _, d := range dates {
		add(d)
		for _, s := range commonSuffixes[:8] {
			add(d + s)
			add(s + d)
		}
	}

	// 5. Leet speak
	if opts.IncludeLeet {
		allWords := append(names, other...)
		if len(allWords) > 20 {
			allWords = allWords[:20]
		}
		for _, word := range allWords {
			addAll(leetVariants(word))
		}
	}

	// 6. Prefixes
	allWords := append(names, other...)
	if len(allWords) > 15 {
		allWords = allWords[:15]
	}
	for _, word := range allWords {
		addAll(withPrefixes(word))
	}

	// 7. Reversed
	if opts.IncludeReverse {
		rev := append(names, other...)
		if len(rev) > 10 {
			rev = rev[:10]
		}
		for _, word := range rev {
			r := reverse(word)
			add(r)
			add(r + "123")
			add(r + "!")
			add(r + "1")
		}
	}

	// 8. Combinations
	if opts.IncludeCombinations {
		addAll(combinations(append(names, other...)))
	}

	// 9. Special char endings
	end := append(names, other...)
	if len(end) > 10 {
		end = end[:10]
	}
	for _, word := range end {
		for _, ch := range specialChars {
			add(word + ch)
			if len(word) > 0 {
				add(strings.ToUpper(word[:1])+word[1:] + ch)
			}
		}
	}

	if len(ordered) > opts.MaxSize {
		ordered = ordered[:opts.MaxSize]
	}
	return ordered
}

// Save writes passwords to a file, one per line.
func Save(passwords []string, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create wordlist dir: %w", err)
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create wordlist file: %w", err)
	}
	defer f.Close()
	for _, p := range passwords {
		fmt.Fprintln(f, p)
	}
	return nil
}

// Filename generates a unique wordlist file path.
func Filename(dataDir, target, sessionID string) string {
	safe := strings.NewReplacer(".", "_", "/", "_", ":", "_").Replace(target)
	ts := time.Now().Format("20060102_150405")
	return filepath.Join(dataDir, "wordlists", fmt.Sprintf("wordlist_%s_%s_%s.txt", safe, sessionID, ts))
}

// Load reads a wordlist file into memory.
func Load(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var passwords []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			passwords = append(passwords, line)
		}
	}
	return passwords, nil
}
