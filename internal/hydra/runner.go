// internal/hydra/runner.go
// THC-Hydra subprocess wrapper.
// Processes passwords in batches, integrates with the tracker.
package hydra

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hydria-ai/hydria/internal/tracker"
)

// foundPattern matches a successful Hydra output line.
var foundPattern = regexp.MustCompile(
	`(?i)(?:\[\d+\])?\[[\w\-]+\]\s+host:\s+(\S+)\s+login:\s+(\S+)\s+password:\s+(.+)`,
)

// IsInstalled checks whether the hydra binary is available on PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("hydra")
	return err == nil
}

// ProgressFunc is called after each attempted password.
// tried = total passwords tried so far, total = full list size, current = password just tried.
type ProgressFunc func(tried, total int, current string)

// Result holds the outcome of an attack.
type Result struct {
	Found    bool
	Password string
	Tried    int
	Skipped  int
}

// RunAttack drives a Hydra attack against target/service/username.
func RunAttack(
	target, service, username string,
	passwords []string,
	port int,
	threads, batchSize int,
	onProgress ProgressFunc,
) (Result, error) {
	// Filter already-tried passwords
	untried, skipped := tracker.FilterUntried(target, username, passwords)
	total := len(untried)
	tried := 0

	for batchStart := 0; batchStart < total; batchStart += batchSize {
		end := batchStart + batchSize
		if end > total {
			end = total
		}
		batch := untried[batchStart:end]

		// Write temp wordlist
		tmpFile, err := os.CreateTemp("", "hydria-*.txt")
		if err != nil {
			return Result{Tried: tried, Skipped: skipped}, fmt.Errorf("create temp file: %w", err)
		}
		for _, p := range batch {
			fmt.Fprintln(tmpFile, p)
		}
		tmpFile.Close()

		found, password, batchTried, err := runBatch(
			target, service, username, tmpFile.Name(),
			port, threads, total, tried, onProgress,
		)
		os.Remove(tmpFile.Name())

		if err != nil {
			tracker.RecordAttemptsBulk(target, username, batch, "failed")
			return Result{Tried: tried, Skipped: skipped}, err
		}

		tried += batchTried

		if found {
			tracker.RecordAttempt(target, username, password, "success")
			remaining := make([]string, 0, len(batch))
			for _, p := range batch {
				if p != password {
					remaining = append(remaining, p)
				}
			}
			tracker.RecordAttemptsBulk(target, username, remaining, "failed")
			return Result{Found: true, Password: password, Tried: tried, Skipped: skipped}, nil
		}

		tracker.RecordAttemptsBulk(target, username, batch, "failed")
	}

	return Result{Tried: tried, Skipped: skipped}, nil
}

func runBatch(
	target, service, username, wordlistPath string,
	port, threads, total, triedSoFar int,
	onProgress ProgressFunc,
) (found bool, password string, tried int, err error) {

	args := []string{
		"-l", username,
		"-P", wordlistPath,
		"-t", fmt.Sprintf("%d", threads),
		"-f",  // stop after first success
		"-V",  // verbose: print every attempt
		"-I",  // ignore previous session file
	}
	if port > 0 {
		args = append(args, "-s", fmt.Sprintf("%d", port))
	}
	args = append(args, target, service)

	cmd := exec.Command("hydra", args...)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err != nil {
		return false, "", 0, fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return false, "", 0, fmt.Errorf("start hydra: %w", err)
	}

	var batchTried []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse verbose attempt line: [ATTEMPT] target ... - pass "xxx"
		if idx := strings.Index(line, `pass "`); idx != -1 {
			rest := line[idx+6:]
			if end := strings.Index(rest, `"`); end != -1 {
				currentPass := rest[:end]
				batchTried = append(batchTried, currentPass)
				tried++
				if onProgress != nil {
					onProgress(triedSoFar+tried, total, currentPass)
				}
			}
		}

		// Check for success
		if m := foundPattern.FindStringSubmatch(line); m != nil {
			password = m[3]
			found = true
			cmd.Process.Kill()
			break
		}
	}

	cmd.Wait()
	return found, password, len(batchTried), nil
}
