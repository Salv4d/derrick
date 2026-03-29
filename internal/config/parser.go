package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseConfig(filename string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var config ProjectConfig

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, enhanceYAMLError(data, err)
	}

	return &config, nil
}

func enhanceYAMLError(fileData []byte, originalErr error) error {
	errMsg := originalErr.Error()

	re := regexp.MustCompile(`line (\d+)`)
	matches := re.FindStringSubmatch(errMsg)

	if len(matches) < 2 {
		return originalErr
	}

	lineNum, err := strconv.Atoi(matches[1])
	if err != nil {
		return originalErr
	}

	lines := strings.Split(string(fileData), "\n")

	if lineNum < 1 || lineNum > len(lines) {
		return originalErr
	}

	errorLine := lines[lineNum-1]

	trimmedLine := strings.TrimSpace(errorLine)
	indentLength := len(errorLine) - len(trimmedLine)
	indicator := strings.Repeat(" ", indentLength) + strings.Repeat("^", len(trimmedLine))

	var builder strings.Builder
	fmt.Fprintf(&builder, "Syntax error in derrick.yaml at line %d: \n\n", lineNum)

	if lineNum > 1 {
		fmt.Fprintf(&builder, "  %3d | %s\n", lineNum-1, lines[lineNum-2])
	}

	fmt.Fprintf(&builder, "  %3d | %s\n", lineNum, errorLine)
	fmt.Fprintf(&builder, "       %s\n\n", indicator)

	fmt.Fprintf(&builder, "Detail: %s", errMsg)

	return errors.New(builder.String())
}