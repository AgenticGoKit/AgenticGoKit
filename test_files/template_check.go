package main

import (
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/templates"
)

func main() {
	fmt.Println("Testing template access...")
	fmt.Printf("WebUIIndexTemplate length: %d\n", len(templates.WebUIIndexTemplate))
	fmt.Printf("WebUICSSTemplate length: %d\n", len(templates.WebUICSSTemplate))

	// Try to access WebUIJSTemplate
	jsTemplate := templates.WebUIJSTemplate
	fmt.Printf("WebUIJSTemplate length: %d\n", len(jsTemplate))
	fmt.Println("Template access successful!")
}
