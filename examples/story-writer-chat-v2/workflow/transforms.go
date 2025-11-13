package workflow

import (
	"fmt"
	"strings"
)

// WriterTransform prepares input for the Writer agent
// Handles both initial draft and revision scenarios
func WriterTransform(input string) string {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("[DEBUG] AGENT: WRITER\n")

	fmt.Printf("[DEBUG WRITER INPUT] Iteration type: ")
	// Check if this is revision with fixes from editor
	if strings.Contains(input, "FIX:") {
		fmt.Printf("REVISION (applying fixes from memory)\n")
		fmt.Printf("[DEBUG WRITER INPUT] Fixes to apply:\n%s\n", input)
		// Editor sent only fixes - Writer retrieves story from memory
		return input + "\n\n[RETRIEVE YOUR PREVIOUS STORY FROM MEMORY AND APPLY THESE FIXES. OUTPUT ONLY THE CORRECTED STORY.]"
	}
	// First iteration - write new story with intentional errors
	fmt.Printf("FIRST DRAFT\n")
	fmt.Printf("[DEBUG WRITER INPUT] User prompt:\n%s\n", input)
	return input + "\n\n[WRITE STORY WITH 2-3 SPELLING ERRORS. OUTPUT ONLY STORY.]"
}

// EditorTransform prepares input for the Editor agent
// Adds instructions for checking spelling and output format
func EditorTransform(input string) string {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("[DEBUG] AGENT: EDITOR\n")

	fmt.Printf("[DEBUG EDITOR INPUT] Story to review:\n%s\n", input)
	fmt.Printf("[DEBUG EDITOR INPUT] Story length: %d chars\n", len(input))

	// Add instruction - Editor only sends FIX list
	return input + "\n\n[CHECK SPELLING. IF ERRORS: Output 'FIX: w1→w2, w3→w4' (NO STORY). IF PERFECT: Output 'APPROVED: [full story]']"
}

// PublisherTransform prepares input for the Publisher agent
// Adds enforcement for format-only output
func PublisherTransform(input string) string {
	// Add enforcement suffix to publisher's prompt
	return input + "\n\n[OUTPUT REQUIREMENT: Format the story with ## Title and paragraphs. Output ONLY the formatted story. No commentary like 'What a delightful story' or 'If you'd like to expand'. Just the formatted story.]"
}
