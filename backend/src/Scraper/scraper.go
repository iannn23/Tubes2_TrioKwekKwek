// Modified scraper.go
package main

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
    Name     string     `json:"name"`
    Recipes  [][]string `json:"recipes"`
    ImageURL string     `json:"imageUrl"` // URL to element's image
}

type ElementGroup struct {
    TierNum  int       `json:"tierNum"`
    Elements []Element `json:"elements"`
}

func ScrapeElements() ([]ElementGroup, error) {
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

    // Maps for storing recipes and element metadata
    recipeMap := make(map[string][][]string)
    imageURLMap := make(map[string]string)      // Maps element name to image URL
    tierNumMap := make(map[string]int)          // Maps element name to numerical tier

    // Extract tiers by tracking headers and tables
    currentTierNum := -1
    
    doc.Find("h2, h3, table.list-table").Each(func(i int, s *goquery.Selection) {
        // Check if it's a heading (h2 or h3)
        if s.Is("h2") || s.Is("h3") {
            headingText := cleanText(s.Text())
            log.Printf("Found heading: %s", headingText)
            
            // Parse tier from heading
            if strings.Contains(strings.ToLower(headingText), "starting elements") {
                currentTierNum = 0
                log.Printf("Current tier: %d", currentTierNum)
            } else if strings.Contains(strings.ToLower(headingText), "special element") {
                currentTierNum = 0 // Special is also considered base level
                log.Printf("Current tier: %d", currentTierNum)
            } else if strings.Contains(strings.ToLower(headingText), "tier") {
                // Extract tier number from heading text
                tierParts := strings.Split(headingText, " ")
                if len(tierParts) >= 2 {
                    // Try to parse tier number
                    tierNumStr := strings.TrimSpace(tierParts[1])
                    switch tierNumStr {
                    case "1":
                        currentTierNum = 1
                    case "2":
                        currentTierNum = 2
                    case "3":
                        currentTierNum = 3
                    case "4":
                        currentTierNum = 4
                    case "5":
                        currentTierNum = 5
                    case "6":
                        currentTierNum = 6
                    case "7":
                        currentTierNum = 7
                    case "8":
                        currentTierNum = 8
                    case "9":
                        currentTierNum = 9
                    case "10":
                        currentTierNum = 10
                    case "11":
                        currentTierNum = 11
                    case "12":
                        currentTierNum = 12
                    case "13":
                        currentTierNum = 13
                    case "14":
                        currentTierNum = 14
                    case "15":
                        currentTierNum = 15
                    default:
                        currentTierNum = -1
                    }
                }
                log.Printf("Current tier: %d", currentTierNum)
            }
        } else if s.Is("table.list-table") && currentTierNum != -1 {
            // Process table with the current tier context
            log.Printf("Processing table for tier: %d", currentTierNum)
            
            // Process rows
            s.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
                // Skip header row
                if row.Find("th").Length() > 0 {
                    return
                }

                cells := row.Find("td")
                if cells.Length() < 2 {
                    return
                }

                // First cell contains the element name and image
                elementCell := cells.Eq(0)
                
                // Extract element name from the last link in the cell
                elementNode := elementCell.Find("a").Last()
                element := cleanText(elementNode.Text())
                
                // Extract image URL from the img tag - use wikia standard URL format
                imgSrc := ""
                elementCell.Find("img").Each(func(i int, img *goquery.Selection) {
                    if src, exists := img.Attr("src"); exists && imgSrc == "" {
                        imgSrc = src
                    } else if dataSrc, exists := img.Attr("data-src"); exists && imgSrc == "" {
                        imgSrc = dataSrc
                    }
                })

                // If we couldn't get a real image, create a URL in the required format
                if imgSrc == "" {
                    elementName := strings.Replace(element, " ", "_", -1)
                    imgSrc = fmt.Sprintf("https://static.wikia.nocookie.net/little-alchemy/images/placeholder/%s_2.svg/revision/latest?cb=20210827000000", elementName)
                }
                
                if element == "" {
                    return
                }

                // Assign tier info and image URL to element
                tierNumMap[element] = currentTierNum
                imageURLMap[element] = imgSrc
                
                log.Printf("Element %q assigned tier %d, image: %s", element, currentTierNum, imgSrc)

                // Extract recipes from the second cell
                recipesCell := cells.Eq(1)
                hasRecipes := false
                
                // Check if "Available from the start" text exists
                if strings.Contains(recipesCell.Text(), "Available from the start") {
                    // This is a starting element with no recipe
                    recipeMap[element] = [][]string{}
                    hasRecipes = true
                }
                
                // Process regular recipes found in list items
                recipesCell.Find("ul li").Each(func(k int, li *goquery.Selection) {
                    // Extract ingredients from <a> tags
                    var recipe []string
                    li.Find("a").Each(func(l int, a *goquery.Selection) {
                        ing := cleanText(a.Text())
                        if ing != "" {
                            recipe = append(recipe, ing)
                        }
                    })

                    // Only accept recipes with exactly 2 ingredients
                    if len(recipe) == 2 {
                        recipeMap[element] = append(recipeMap[element], recipe)
                        hasRecipes = true
                        log.Printf("Added recipe for %q: %v (tier: %d)", element, recipe, currentTierNum)
                    }
                })

                // If no recipes were found but text mentions unlocking requirements
                if !hasRecipes && strings.Contains(recipesCell.Text(), "unlock") {
                    // Special elements like Time
                    recipeMap[element] = [][]string{}
                    hasRecipes = true
                    log.Printf("Added special element %q with unlock requirement", element)
                }

                // Add element even if it has no recipes
                if !hasRecipes && !recipeMapContainsElement(recipeMap, element) {
                    recipeMap[element] = [][]string{}
                    log.Printf("Added element %q with no recipes (tier: %d)", element, currentTierNum)
                }
            })
        }
    })

    // Group elements by tier
    tierGroups := make(map[int][]Element)
    
    // Convert maps to Element structs and group by tier
    for name, recipes := range recipeMap {
        // Sort ingredients within each recipe
        for _, recipe := range recipes {
            sort.Strings(recipe)
        }
        
        // Sort recipes by first ingredient
        sort.Slice(recipes, func(i, j int) bool {
            if len(recipes[i]) == 0 || len(recipes[j]) == 0 {
                return len(recipes[i]) < len(recipes[j])
            }
            return recipes[i][0] < recipes[j][0]
        })

        // Get tierNum and imageURL, default to -1 and empty string if not found
        tierNum, exists := tierNumMap[name]
        if !exists {
            tierNum = -1
        }
        
        imageURL, exists := imageURLMap[name]
        if !exists {
            // Use standard format if missing
            imgName := strings.Replace(name, " ", "_", -1)
            imageURL = fmt.Sprintf("https://static.wikia.nocookie.net/little-alchemy/images/placeholder/%s.svg/revision/latest/scale-to-width-down/40", imgName)
        }

        // Create element and add to appropriate tier group
        element := Element{
            Name:     name,
            Recipes:  recipes,
            ImageURL: imageURL,
        }
        
        tierGroups[tierNum] = append(tierGroups[tierNum], element)
    }

    // Sort tiers and elements within tiers
    var result []ElementGroup
    
    // Get sorted list of tier numbers
    tiers := make([]int, 0, len(tierGroups))
    for tier := range tierGroups {
        tiers = append(tiers, tier)
    }
    sort.Ints(tiers)
    
    // Create element groups in order
    for _, tier := range tiers {
        elements := tierGroups[tier]
        // Sort elements by name within each tier
        sort.Slice(elements, func(i, j int) bool {
            return elements[i].Name < elements[j].Name
        })
        
        result = append(result, ElementGroup{
            TierNum:  tier,
            Elements: elements,
        })
    }

    if len(result) == 0 {
        return nil, fmt.Errorf("no elements found")
    }

    // Print tier statistics
    for _, group := range result {
        log.Printf("Tier %d: %d elements", group.TierNum, len(group.Elements))
    }

    return result, nil
}

func recipeMapContainsElement(recipeMap map[string][][]string, element string) bool {
    _, exists := recipeMap[element]
    return exists
}

// cleanText normalizes text by removing extra whitespace, newlines, and tabs
func cleanText(s string) string {
    s = strings.TrimSpace(s)
    s = strings.ReplaceAll(s, "\n", " ")
    s = strings.ReplaceAll(s, "\t", " ")
    for strings.Contains(s, "  ") {
        s = strings.ReplaceAll(s, "  ", " ")
    }
    return s
}