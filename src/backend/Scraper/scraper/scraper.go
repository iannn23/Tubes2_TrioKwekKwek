package scraper

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Element struct {
	Name    string     `json:"name"`
	Recipes [][]string `json:"recipes"`
}

func ScrapeElements() ([]Element, error) {
	const url = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

	// Create HTTP client with timeout and custom headers
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching URL %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	recipeMap := make(map[string][][]string)

	// Process all tables with class "list-table"
	tables := doc.Find("table.list-table")
	log.Printf("Found %d tables with class 'list-table'", tables.Length())
	tables.Each(func(tableIdx int, table *goquery.Selection) {
		log.Printf("Processing table %d", tableIdx)
		rows := table.Find("tbody tr")
		log.Printf("Table %d: Found %d rows", tableIdx, rows.Length())

		rows.Each(func(i int, s *goquery.Selection) {
			// Skip header row (contains <th> instead of <td>)
			if s.Find("th").Length() > 0 {
				log.Printf("Table %d, Row %d: Skipped (header row)", tableIdx, i)
				return
			}

			cells := s.Find("td")
			log.Printf("Table %d, Row %d: %d cells", tableIdx, i, cells.Length())

			if cells.Length() < 2 {
				log.Printf("Table %d, Row %d: Skipped (fewer than 2 cells)", tableIdx, i)
				return
			}

			// Element name (first column, inside <a> tag)
			elementNode := cells.Eq(0).Find("a").Last() // Get the last <a> (the one with the element name)
			element := cleanText(elementNode.Text())
			log.Printf("Table %d, Row %d: Element=%q", tableIdx, i, element)

			if element == "" {
				log.Printf("Table %d, Row %d: Skipped (empty element)", tableIdx, i)
				return
			}

			// Recipes (second column, inside <ul><li>)
			recipesCell := cells.Eq(1)
			hasRecipes := false
			recipesCell.Find("ul li").Each(func(j int, li *goquery.Selection) {
				// Extract ingredients by splitting on " + "
				recipeText := li.Text()
				ingredients := strings.Split(recipeText, " + ")
				log.Printf("Table %d, Row %d, Recipe %d: Raw ingredients=%v", tableIdx, i, j, ingredients)

				if len(ingredients) != 2 {
					log.Printf("Table %d, Row %d, Recipe %d: Skipped (invalid ingredient count: %d)", tableIdx, i, j, len(ingredients))
					return
				}

				// Clean each ingredient by extracting text from <a> tags
				var recipe []string
				li.Find("a").Each(func(k int, a *goquery.Selection) {
					ing := cleanText(a.Text())
					if ing != "" {
						recipe = append(recipe, ing)
					}
				})

				if len(recipe) == 2 {
					log.Printf("Table %d, Row %d, Recipe %d: Valid recipe: %v", tableIdx, i, j, recipe)
					recipeMap[element] = append(recipeMap[element], recipe)
					hasRecipes = true
				} else {
					log.Printf("Table %d, Row %d, Recipe %d: Skipped (invalid recipe length: %d)", tableIdx, i, j, len(recipe))
				}
			})

			// If no recipes were found, still add the element with an empty recipes list
			if !hasRecipes {
				log.Printf("Table %d, Row %d: Element %q has no recipes", tableIdx, i, element)
				if _, exists := recipeMap[element]; !exists {
					recipeMap[element] = [][]string{}
				}
			}
		})
	})

	// Convert map to slice and sort
	var elements []Element
	for name, recipes := range recipeMap {
		// Sort ingredients within each recipe
		for _, recipe := range recipes {
			sort.Strings(recipe)
		}
		// Sort recipes by first ingredient
		sort.Slice(recipes, func(i, j int) bool {
			return recipes[i][0] < recipes[j][0]
		})

		elements = append(elements, Element{
			Name:    name,
			Recipes: recipes,
		})
	}

	// Sort elements by name
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Name < elements[j].Name
	})

	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found")
	}

	log.Printf("Found %d elements", len(elements))
	return elements, nil
}

// cleanText normalizes text by removing extra whitespace, newlines, and tabs
func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "  ", " ")
	return s
}
