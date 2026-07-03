package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: coverage-filter <exclude-file> <coverage-profile>\n")
		os.Exit(1)
	}

	excludeFile := os.Args[1]
	profileFile := os.Args[2]

	excludes, err := loadExcludes(excludeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading excludes: %v\n", err)
		os.Exit(1)
	}

	profile, err := os.Open(profileFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening profile: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = profile.Close() }()

	scanner := bufio.NewScanner(profile)
	totalLines := 0
	excludedLines := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "mode:") {
			fmt.Println(line)
			continue
		}

		if isExcluded(line, excludes) {
			excludedLines++
			continue
		}

		fmt.Println(line)
		totalLines++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Excluded %d blocks from coverage profile\n", excludedLines)
}

func loadExcludes(path string) (map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	excludes := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		excludes[line] = true
	}
	return excludes, scanner.Err()
}

func isExcluded(line string, excludes map[string]bool) bool {
	for exclude := range excludes {
		if strings.HasPrefix(line, exclude+" ") {
			return true
		}
	}
	return false
}
