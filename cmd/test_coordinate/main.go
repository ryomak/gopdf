package main

import (
	"fmt"
	"log"

	"github.com/ryomak/gopdf"
)

func main() {
	if err := gopdf.TestCoordinateIssue(); err != nil {
		log.Fatalf("Test failed: %v", err)
	}
	fmt.Println("\nTest completed successfully!")
}
