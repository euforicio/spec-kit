package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for a yes/no confirmation
func Confirm(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

// PromptString prompts the user for a string input
func PromptString(message string) string {
	fmt.Printf("%s", message)
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}
	
	return strings.TrimSpace(scanner.Text())
}

// PromptStringWithDefault prompts the user for a string input with a default value
func PromptStringWithDefault(message, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", message, defaultValue)
	} else {
		fmt.Printf("%s: ", message)
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return defaultValue
	}
	
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultValue
	}
	
	return input
}

// PromptSelect prompts the user to select from a list of options
func PromptSelect(message string, options []string) string {
	for {
		input := PromptString(message)
		
		// Check if input is valid
		for _, option := range options {
			if input == option {
				return input
			}
		}
		
		fmt.Printf("Invalid choice. Please select from: %s\n", strings.Join(options, ", "))
	}
}

// PromptSelectIndex prompts the user to select from a list and returns the index
func PromptSelectIndex(message string, options []string) int {
	if len(options) == 0 {
		return -1
	}
	
	// Display options
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}
	fmt.Println()
	
	for {
		input := PromptString(message)
		
		// Try to parse as number
		for i := range options {
			if input == fmt.Sprintf("%d", i+1) {
				return i
			}
		}
		
		fmt.Printf("Invalid choice. Please enter a number between 1 and %d\n", len(options))
	}
}

// PromptPassword prompts the user for a password (input is hidden)
// Note: This is a basic implementation. For production use, consider using
// a library like golang.org/x/term for proper password input handling
func PromptPassword(message string) string {
	fmt.Printf("%s: ", message)
	
	// This is a simplified version - doesn't actually hide input
	// In a real implementation, you'd use golang.org/x/term
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}
	
	return strings.TrimSpace(scanner.Text())
}

// PromptMultiSelect prompts the user to select multiple options
func PromptMultiSelect(message string, options []string) []string {
	selected := make([]string, 0)
	
	fmt.Printf("%s (enter numbers separated by commas, or 'done' to finish):\n", message)
	
	// Display options
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}
	fmt.Println()
	
	for {
		input := PromptString("Select: ")
		
		if strings.ToLower(input) == "done" {
			break
		}
		
		// Parse comma-separated numbers
		parts := strings.Split(input, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			
			for i := range options {
				if part == fmt.Sprintf("%d", i+1) {
					// Check if already selected
					alreadySelected := false
					for _, sel := range selected {
						if sel == options[i] {
							alreadySelected = true
							break
						}
					}
					
					if !alreadySelected {
						selected = append(selected, options[i])
						fmt.Printf("Added: %s\n", options[i])
					}
				}
			}
		}
		
		if len(selected) > 0 {
			fmt.Printf("Selected: %s\n", strings.Join(selected, ", "))
		}
	}
	
	return selected
}

// ShowProgress displays a simple progress indicator
func ShowProgress(current, total int, message string) {
	if total <= 0 {
		return
	}
	
	percentage := float64(current) / float64(total) * 100
	
	// Simple progress bar
	barWidth := 40
	filled := int(float64(barWidth) * float64(current) / float64(total))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	
	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", message, bar, percentage, current, total)
	
	if current >= total {
		fmt.Println() // New line when complete
	}
}

// ClearLine clears the current line in the terminal
func ClearLine() {
	fmt.Print("\r\033[K")
}

// MoveCursorUp moves the cursor up by the specified number of lines
func MoveCursorUp(lines int) {
	fmt.Printf("\033[%dA", lines)
}

// MoveCursorDown moves the cursor down by the specified number of lines
func MoveCursorDown(lines int) {
	fmt.Printf("\033[%dB", lines)
}