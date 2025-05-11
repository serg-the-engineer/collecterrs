package main

import (
	"collecterrs/collecterrs"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Get the module name from the project's go.mod file
	moduleName, err := getModuleName("project/go.mod")
	if err != nil {
		fmt.Printf("error getting module name: %v\n", err)
		return
	}

	ua := collecterrs.NewUsecaseAnalysis()
	results, err := ua.Analyze("project/services", moduleName, false)
	if err != nil {
		fmt.Printf("error analyzing usecases: %v\n", err)
		return
	}

	output, _ := json.MarshalIndent(results, "", "  ")

	// Save the output to a file
	err = os.WriteFile("project-errors.json", output, 0644)
	if err != nil {
		fmt.Printf("error writing to file: %v\n", err)
		return
	}

	fmt.Println("Results saved to project-errors.json")
}

// getModuleName reads the module name from a go.mod file
func getModuleName(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod file")
}
