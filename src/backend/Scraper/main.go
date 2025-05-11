package main

import (
	"encoding/json"
	"fmt"
	"os"

	"Scraper/scraper"
)

func main() {
	elements, err := scraper.ScrapeElements()
	if err != nil {
		fmt.Println("Gagal scrape:", err)
		return
	}

	file, _ := os.Create("elements.json")
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(elements)

	fmt.Printf("Scraping selesai: %d elemen disimpan ke elements.json\n", len(elements))
}
