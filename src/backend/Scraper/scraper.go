package Scraper

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

func ScrapeElements() {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch page: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Non-200 response: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	result := make(map[string][][]string)

	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		// Elemen biasanya dalam tag <h3><span class="mw-headline" id="ELEMENT_NAME">
		elemenNama := s.Find("span.mw-headline").Text()
		if elemenNama == "" {
			return
		}

		// Cari resep yang muncul tepat setelah elemen
		var recipes [][]string
		next := s.Next()
		for next.Is("ul") {
			next.Find("li").Each(func(i int, li *goquery.Selection) {
				text := strings.TrimSpace(li.Text())
				// Format umum: A + B = ELEMENT_NAME
				parts := strings.Split(text, " = ")
				if len(parts) != 2 {
					return
				}
				ingredients := strings.Split(parts[0], " + ")
				if len(ingredients) == 2 {
					recipes = append(recipes, ingredients)
				}
			})
			next = next.Next()
		}

		if len(recipes) > 0 {
			result[elemenNama] = recipes
		}
	})

	// Simpan ke file JSON
	file, err := os.Create("data/elements.json")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	log.Println("Scraping selesai. File disimpan di data/elements.json")
}
