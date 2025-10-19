package main

import (
	"fmt"
	"os"

	"spacectl/internal/models"
	"spacectl/internal/output"
)

func main() {
	// Create a test project
	project := models.Project{
		ID:             "c9cfaf99-949c-45f3-bd08-442c1064fe06",
		Name:           "test",
		OrganizationID: "215e8c81-6d9c-4175-ba82-428f2b0818d3",
		MaxTenants:     0,
		MaxCompute:     0,
		MaxMemoryGB:    0,
	}

	// Create formatter
	formatter := output.NewFormatter(output.FormatTable, false, os.Stdout)

	// Test the formatting
	fmt.Println("Testing Project formatting:")
	fmt.Println("==========================")
	err := formatter.FormatData(project)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
