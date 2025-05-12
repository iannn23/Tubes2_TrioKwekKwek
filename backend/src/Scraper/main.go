package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
)

func main() {
    elements, err := ScrapeElements()  // Direct call, no package name needed
    if err != nil {
        fmt.Println("Failed scraping:", err)
        return
    }

    file, err := os.Create("elements.json")
    if err != nil {
        log.Fatalf("Error creating file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(elements); err != nil {
        log.Fatalf("Error encoding JSON: %v", err)
    }

    fmt.Printf("Scraping complete: %d tier groups saved to elements.json\n", len(elements))
}