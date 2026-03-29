package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})
}

func ParseConfig(filename string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var config ProjectConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	err = decoder.Decode(&config)
	if err != nil {
		return nil, enhanceYAMLError(data, err)
	}

	err = validate.Struct(config)
	if err != nil {
		return nil, formatValidationError(err)
	}

	return &config, nil
}

func formatValidationError(err error) error {
	var builder strings.Builder
	builder.WriteString("Invalid configuration contract in derrick.yaml:\n\n")

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			yamlPath := strings.TrimPrefix(e.Namespace(), "ProjectConfig.")
			fmt.Fprintf(&builder, "  ✖ Field '%s' failed validation: must be '%s'\n", yamlPath, e.Tag())
		}
		return errors.New(builder.String())
	}

	return err
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