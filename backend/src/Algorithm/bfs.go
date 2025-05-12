package main

import (
    "container/list"
    "fmt"
    "sync"
    "time"
)

// RecipeStep for path tracking
type RecipeStep struct {
    ParentID string
    Recipe   Recipe
}

// BreadthFirstFinder for recipe search
type BreadthFirstFinder struct {
    store *ElementStore
}

// NewBreadthFirstFinder creates finder instance
func NewBreadthFirstFinder(store *ElementStore) *BreadthFirstFinder {
    return &BreadthFirstFinder{store: store}
}

// FindShortestPath finds shortest recipe path
func (bf *BreadthFirstFinder) FindShortestPath(target string) (*SearchResult, error) {
    startTime := time.Now()

    // Check target exists
    _, exists := bf.store.Elements[target]
    if !exists {
        return nil, ErrElementNotFound
    }

    // Get basic elements
    basicElements := bf.store.GetBasicElements()
    if len(basicElements) == 0 {
        return nil, ErrNoBasicElements
    }

    // Count visited nodes
    visitedCount := 0

    // BFS setup
    queue := list.New()
    visited := make(map[string]bool)
    parent := make(map[string]RecipeStep) // Parent tracking

    // Initialize queue with basic elements
    for _, elem := range basicElements {
        queue.PushBack(elem.ID)
        visited[elem.ID] = true
        visitedCount++
    }

    // Path found flag
    found := false

    // Run BFS
    for queue.Len() > 0 && !found {
        current := queue.Remove(queue.Front()).(string)
        currentTier := bf.store.GetElementTier(current)

        // Check if we found the target
        if current == target {
            found = true
            break
        }

        // Expand current node - only consider recipes that produce higher tier elements
        possibleRecipes := bf.getPossibleRecipesThatRespectTiers(current, currentTier)
        for _, recipe := range possibleRecipes {
            resultElem := recipe.Result
            resultTier := bf.store.GetElementTier(resultElem)
            
            // Only consider recipes that lead to higher tiers than current
            if resultTier <= currentTier {
                continue
            }
            
            if !visited[resultElem] {
                queue.PushBack(resultElem)
                visited[resultElem] = true
                parent[resultElem] = RecipeStep{
                    ParentID: current,
                    Recipe:   recipe,
                }
                visitedCount++

                // Early exit if we found the target
                if resultElem == target {
                    found = true
                    break
                }
            }
        }
    }

    if !found {
        return nil, ErrNoPathFound
    }

    // Build path
    path := bf.reconstructPath(target, parent)

    // Visualize tree
    treeStructure := bf.buildTreeStructure(path, target)

    executionTime := time.Since(startTime).Milliseconds()

    return &SearchResult{
        Path:          path,
        VisitedNodes:  visitedCount,
        ExecutionTime: executionTime,
        TreeStructure: treeStructure,
    }, nil
}

// FindMultiplePaths finds multiple recipe paths
func (bf *BreadthFirstFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target exists
    _, exists := bf.store.Elements[target]
    if !exists {
        return nil, ErrElementNotFound
    }

    var results []*SearchResult
    var mutex sync.Mutex
    var wg sync.WaitGroup

    // Limit goroutines
    maxGoroutines := 4
    if maxGoroutines > maxPaths {
        maxGoroutines = maxPaths
    }

    // Concurrency control
    sem := make(chan bool, maxGoroutines)

    // Start multiple searches
    for i := 0; i < maxPaths; i++ {
        wg.Add(1)
        sem <- true

        go func(index int) {
            defer wg.Done()
            defer func() { <-sem }()

            // Find variation path
            result, err := bf.findPathWithVariation(target, index)

            if err == nil {
                mutex.Lock()
                results = append(results, result)
                mutex.Unlock()
            }
        }(i)
    }

    wg.Wait()

    // Check results
    if len(results) == 0 {
        return nil, ErrNoPathFound
    }

    return results, nil
}

// Get recipes using element that respect tier hierarchy
func (bf *BreadthFirstFinder) getPossibleRecipesThatRespectTiers(elementID string, currentTier int) []Recipe {
    var recipes []Recipe

    for _, recipe := range bf.store.Recipes {
        for _, ingredient := range recipe.Ingredients {
            if ingredient == elementID {
                // Check if the result tier is higher than current tier
                resultTier := bf.store.GetElementTier(recipe.Result)
                if resultTier > currentTier {
                    recipes = append(recipes, recipe)
                }
                break
            }
        }
    }

    return recipes
}

// Get all recipes using element (without tier restriction)
func (bf *BreadthFirstFinder) getPossibleRecipesWith(elementID string) []Recipe {
    var recipes []Recipe

    for _, recipe := range bf.store.Recipes {
        for _, ingredient := range recipe.Ingredients {
            if ingredient == elementID {
                recipes = append(recipes, recipe)
                break
            }
        }
    }

    return recipes
}

// Build path
func (bf *BreadthFirstFinder) reconstructPath(target string, parentMap map[string]RecipeStep) []Recipe {
    var path []Recipe
    current := target
    
    // Trace path from target back to a basic element
    for {
        step, exists := parentMap[current]
        if !exists {
            break // Reached a basic element
        }
        
        path = append([]Recipe{step.Recipe}, path...) // Prepend to maintain order
        current = step.ParentID
    }
    
    return path
}

// Find variant path
func (bf *BreadthFirstFinder) findPathWithVariation(target string, variationIndex int) (*SearchResult, error) {
    result, err := bf.FindShortestPath(target)

    if err == nil && result != nil {
        // Add variation marker
        result.VariationIndex = variationIndex
    }

    return result, err
}

// Create visualization tree
func (bf *BreadthFirstFinder) buildTreeStructure(path []Recipe, target string) interface{} {
    // Build tree for Next.js visualization - target as root, basics as leaves
    nodes := []map[string]interface{}{}
    edges := []map[string]interface{}{}
    nodeMap := make(map[string]bool)
    
    // Map recipes by result for faster lookup
    recipesByResult := make(map[string]Recipe)
    for _, recipe := range path {
        recipesByResult[recipe.Result] = recipe
    }
    
    // Add target node
    targetElem := bf.store.Elements[target]
    targetTier := bf.store.GetElementTier(target)
    nodes = append(nodes, map[string]interface{}{
        "id":       target,
        "label":    target,
        "type":     "target",
        "tier":     targetTier,
        "imageUrl": targetElem.ImageURL,
    })
    nodeMap[target] = true
    
    // Helper function to recursively build the tree from target to basics
    var buildNodes func(element string)
    buildNodes = func(element string) {
        recipe, exists := recipesByResult[element]
        if !exists {
            // This is a basic element or we've reached a leaf
            return
        }
        
        // Add ingredients as nodes and connect them with edges
        for _, ingredient := range recipe.Ingredients {
            if !nodeMap[ingredient] {
                ingElem := bf.store.Elements[ingredient]
                ingTier := bf.store.GetElementTier(ingredient)
                
                ingType := "ingredient"
                if bf.store.IsBasicElement(ingredient) {
                    ingType = "basic" // Mark basic elements
                }
                
                nodes = append(nodes, map[string]interface{}{
                    "id":       ingredient,
                    "label":    ingredient,
                    "type":     ingType,
                    "tier":     ingTier,
                    "imageUrl": ingElem.ImageURL,
                })
                nodeMap[ingredient] = true
            }
            
            // Add edge from ingredient to result element
            edges = append(edges, map[string]interface{}{
                "id":     fmt.Sprintf("%s-%s", ingredient, element),
                "source": ingredient,
                "target": element,
            })
            
            // Recursively build tree for this ingredient
            buildNodes(ingredient)
        }
    }
    
    // Start building from target
    buildNodes(target)
    
    return map[string]interface{}{
        "nodes":   nodes,
        "edges":   edges,
        "target":  target,
        "recipes": path,
    }
}