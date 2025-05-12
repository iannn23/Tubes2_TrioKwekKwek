package main

import (
    "fmt"
    "sync"
    "time"
)

// DepthFirstFinder for recipe search
type DepthFirstFinder struct {
    store *ElementStore
}

// NewDepthFirstFinder creates finder instance
func NewDepthFirstFinder(store *ElementStore) *DepthFirstFinder {
    return &DepthFirstFinder{store: store}
}

// FindShortestPath finds shortest recipe path using DFS
func (df *DepthFirstFinder) FindShortestPath(target string) (*SearchResult, error) {
    startTime := time.Now()

    // Check target exists
    _, exists := df.store.Elements[target]
    if !exists {
        return nil, ErrElementNotFound
    }
    targetTier := df.store.GetElementTier(target)

    // Get basic elements
    basicElements := df.store.GetBasicElements()
    if len(basicElements) == 0 {
        return nil, ErrNoBasicElements
    }

    // Visited nodes counter
    visitedCount := 0
    
    // Set up DFS
    visited := make(map[string]bool)
    parent := make(map[string]RecipeStep)
    
    // Path found flag
    found := false
    
    // Initialize with depth limit to prevent infinite recursion
    maxDepth := targetTier * 2 // Heuristic: maximum depth is twice the target tier
    
    // Try each basic element as a starting point
    for _, elem := range basicElements {
        visited[elem.ID] = true
        visitedCount++
        
        // Run DFS with depth limit and tier constraints
        found = df.dfsSearchWithTiers(elem.ID, target, visited, parent, 0, maxDepth, &visitedCount)
        
        if found {
            break
        }
        
        // Reset for next basic element
        delete(visited, elem.ID)
    }
    
    if !found {
        return nil, ErrNoPathFound
    }
    
    // Build path
    path := df.reconstructPath(target, parent)
    
    // Visualize tree
    treeStructure := df.buildTreeStructure(path, target)
    
    executionTime := time.Since(startTime).Milliseconds()
    
    return &SearchResult{
        Path:          path,
        VisitedNodes:  visitedCount,
        ExecutionTime: executionTime,
        TreeStructure: treeStructure,
    }, nil
}

// DFS search with depth limit and tier constraints
func (df *DepthFirstFinder) dfsSearchWithTiers(current, target string, visited map[string]bool, 
    parent map[string]RecipeStep, depth, maxDepth int, visitedCount *int) bool {
    
    // Check if we found the target
    if current == target {
        return true
    }
    
    // Check depth limit to prevent infinite recursion
    if depth >= maxDepth {
        return false
    }
    
    // Get current tier
    currentTier := df.store.GetElementTier(current)
    
    // Get possible recipes using current element that lead to higher tiers
    possibleRecipes := df.getPossibleRecipesThatRespectTiers(current, currentTier)
    
    // Try each recipe
    for _, recipe := range possibleRecipes {
        resultElem := recipe.Result
        resultTier := df.store.GetElementTier(resultElem)
        
        // Only consider recipes that lead to higher tiers
        if resultTier <= currentTier {
            continue
        }
        
        if !visited[resultElem] {
            // Mark as visited
            visited[resultElem] = true
            *visitedCount++
            
            // Record parent
            parent[resultElem] = RecipeStep{
                ParentID: current,
                Recipe:   recipe,
            }
            
            // Recurse deeper
            if df.dfsSearchWithTiers(resultElem, target, visited, parent, depth+1, maxDepth, visitedCount) {
                return true
            }
            
            // Backtrack if needed
            delete(visited, resultElem)
            delete(parent, resultElem)
        }
    }
    
    return false
}

// FindMultiplePaths finds multiple recipe paths
func (df *DepthFirstFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target exists
    _, exists := df.store.Elements[target]
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

    // Start multiple searches with different depth limits
    for i := 0; i < maxPaths; i++ {
        wg.Add(1)
        sem <- true

        go func(index int) {
            defer wg.Done()
            defer func() { <-sem }()

            // Find variation path with different depth or starting points
            result, err := df.findPathWithVariation(target, index)

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
func (df *DepthFirstFinder) getPossibleRecipesThatRespectTiers(elementID string, currentTier int) []Recipe {
    var recipes []Recipe

    for _, recipe := range df.store.Recipes {
        for _, ingredient := range recipe.Ingredients {
            if ingredient == elementID {
                // Check if the result tier is higher than current tier
                resultTier := df.store.GetElementTier(recipe.Result)
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
func (df *DepthFirstFinder) getPossibleRecipesWith(elementID string) []Recipe {
    var recipes []Recipe

    for _, recipe := range df.store.Recipes {
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
func (df *DepthFirstFinder) reconstructPath(target string, parentMap map[string]RecipeStep) []Recipe {
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
func (df *DepthFirstFinder) findPathWithVariation(target string, variationIndex int) (*SearchResult, error) {
    // Start search with a different ordering of basic elements or a different depth limit
    startTime := time.Now()

    // Check target exists
    _, exists := df.store.Elements[target]
    if !exists {
        return nil, ErrElementNotFound
    }
    targetTier := df.store.GetElementTier(target)

    // Get basic elements
    basicElements := df.store.GetBasicElements()
    if len(basicElements) == 0 {
        return nil, ErrNoBasicElements
    }

    // Visited nodes counter
    visitedCount := 0
    
    // Set up DFS
    visited := make(map[string]bool)
    parent := make(map[string]RecipeStep)
    
    // Path found flag
    found := false
    
    // Use variation index to adjust depth limit or element priority
    maxDepth := targetTier * 2 + (variationIndex % 5) // Vary max depth
    startElemIndex := variationIndex % len(basicElements) // Vary starting element
    
    // Reorder basic elements based on variation index
    reorderedElements := make([]*Element, len(basicElements))
    copy(reorderedElements, basicElements)
    
    // Rotate elements by startElemIndex
    if startElemIndex > 0 && len(reorderedElements) > 1 {
        reorderedElements = append(reorderedElements[startElemIndex:], reorderedElements[:startElemIndex]...)
    }
    
    // Try each basic element as a starting point with the new ordering
    for _, elem := range reorderedElements {
        visited[elem.ID] = true
        visitedCount++
        
        // Run DFS with depth limit to find shortest path
        found = df.dfsSearchWithTiers(elem.ID, target, visited, parent, 0, maxDepth, &visitedCount)
        
        if found {
            break
        }
        
        // Reset for next basic element
        delete(visited, elem.ID)
    }
    
    if !found {
        return nil, ErrNoPathFound
    }
    
    // Build path
    path := df.reconstructPath(target, parent)
    
    // Visualize tree
    treeStructure := df.buildTreeStructure(path, target)
    
    executionTime := time.Since(startTime).Milliseconds()
    
    result := &SearchResult{
        Path:           path,
        VisitedNodes:   visitedCount,
        ExecutionTime:  executionTime,
        TreeStructure:  treeStructure,
        VariationIndex: variationIndex,
    }
    
    return result, nil
}

// Create visualization tree
func (df *DepthFirstFinder) buildTreeStructure(path []Recipe, target string) interface{} {
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
    targetElem := df.store.Elements[target]
    targetTier := df.store.GetElementTier(target)
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
                ingElem := df.store.Elements[ingredient]
                ingTier := df.store.GetElementTier(ingredient)
                
                ingType := "ingredient"
                if df.store.IsBasicElement(ingredient) {
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