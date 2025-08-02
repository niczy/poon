package merge

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PatchHeader struct {
	OldFile string
	NewFile string
	OldMode string
	NewMode string
}

type PatchHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []PatchLine
}

type PatchLine struct {
	Type    string // "+", "-", " " (context)
	Content string
}

type ParsedPatch struct {
	Header PatchHeader
	Hunks  []PatchHunk
}

func ValidatePatch(patchData []byte) error {
	if len(patchData) == 0 {
		return fmt.Errorf("patch data is empty")
	}

	scanner := bufio.NewScanner(bytes.NewReader(patchData))
	hasValidHeader := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			hasValidHeader = true
		}
		if strings.HasPrefix(line, "@@") {
			if !hasValidHeader {
				return fmt.Errorf("patch has hunk without proper file headers")
			}
		}
	}

	if !hasValidHeader {
		return fmt.Errorf("patch does not contain valid unified diff headers")
	}

	return nil
}

func ParsePatch(patchData []byte) (*ParsedPatch, error) {
	if err := ValidatePatch(patchData); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(patchData))
	patch := &ParsedPatch{}
	var currentHunk *PatchHunk

	hunkRegex := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "--- ") {
			oldFile := strings.TrimPrefix(line, "--- ")
			if strings.HasPrefix(oldFile, "a/") {
				oldFile = oldFile[2:]
			}
			patch.Header.OldFile = oldFile
		} else if strings.HasPrefix(line, "+++ ") {
			newFile := strings.TrimPrefix(line, "+++ ")
			if strings.HasPrefix(newFile, "b/") {
				newFile = newFile[2:]
			}
			patch.Header.NewFile = newFile
		} else if matches := hunkRegex.FindStringSubmatch(line); matches != nil {
			if currentHunk != nil {
				patch.Hunks = append(patch.Hunks, *currentHunk)
			}

			oldStart, _ := strconv.Atoi(matches[1])
			oldCount := 1
			if matches[2] != "" {
				oldCount, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newCount := 1
			if matches[4] != "" {
				newCount, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &PatchHunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
			}
		} else if currentHunk != nil && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ")) {
			patchLine := PatchLine{
				Type:    string(line[0]),
				Content: line[1:],
			}
			currentHunk.Lines = append(currentHunk.Lines, patchLine)
		}
	}

	if currentHunk != nil {
		patch.Hunks = append(patch.Hunks, *currentHunk)
	}

	return patch, nil
}

func BackupFile(filePath string) (string, error) {
	backupPath := filePath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())

	input, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to open file for backup: %v", err)
	}
	defer input.Close()

	output, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy file for backup: %v", err)
	}

	return backupPath, nil
}

func ApplyPatch(filePath string, patch *ParsedPatch) error {
	var originalLines []string

	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %v", err)
		}
		originalContent := string(content)
		if originalContent != "" {
			originalLines = strings.Split(originalContent, "\n")
			if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
				originalLines = originalLines[:len(originalLines)-1]
			}
		}
	}

	result := make([]string, 0, len(originalLines)+100)
	originalIndex := 0

	for _, hunk := range patch.Hunks {
		for originalIndex < hunk.OldStart-1 && originalIndex < len(originalLines) {
			result = append(result, originalLines[originalIndex])
			originalIndex++
		}

		for _, patchLine := range hunk.Lines {
			switch patchLine.Type {
			case " ":
				if originalIndex < len(originalLines) {
					result = append(result, originalLines[originalIndex])
					originalIndex++
				}
			case "-":
				if originalIndex < len(originalLines) {
					originalIndex++
				}
			case "+":
				result = append(result, patchLine.Content)
			}
		}
	}

	for originalIndex < len(originalLines) {
		result = append(result, originalLines[originalIndex])
		originalIndex++
	}

	newContent := strings.Join(result, "\n")
	if len(result) > 0 {
		newContent += "\n"
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write patched file: %v", err)
	}

	return nil
}
