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
    _, exists := df.store.Elements[target]  // Remove targetElem
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
        found = df.dfsSearchWithTiers(elem.ID, target, visited, parent, 0, maxDepth, targetTier, &visitedCount)
        
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
func (df *DepthFirstFinder) dfsSearchWithTiers(
    current, target string, 
    visited map[string]bool, 
    parent map[string]RecipeStep, 
    depth, maxDepth int,
    targetTier int,
    visitedCount *int) bool {
    
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
    possibleRecipes := df.getPossibleRecipesThatRespectTiers(current, currentTier, targetTier)
    
    // Try each recipe
    for _, recipe := range possibleRecipes {
        resultElem := recipe.Result
        resultTier := df.store.GetElementTier(resultElem)
        
        // Skip if any ingredient in the recipe doesn't respect tier constraint
        validRecipe := true
        for _, ingredient := range recipe.Ingredients {
            ingredientTier := df.store.GetElementTier(ingredient)
            if ingredientTier >= resultTier {
                validRecipe = false
                break
            }
        }
        
        if !validRecipe {
            continue
        }
        
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
            if df.dfsSearchWithTiers(resultElem, target, visited, parent, depth+1, maxDepth, targetTier, visitedCount) {
                return true
            }
            
            // Backtrack if needed
            delete(visited, resultElem)
            delete(parent, resultElem)
        }
    }
    
    return false
}

// FindMultiplePaths finds multiple recipe paths using multithreading
func (df *DepthFirstFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target exists
    _, exists := df.store.Elements[target]  // Remove targetElem
    if !exists {
        return nil, ErrElementNotFound
    }
    targetTier := df.store.GetElementTier(target)

    var results []*SearchResult
    var mutex sync.Mutex
    var wg sync.WaitGroup

    // Limit goroutines based on system capabilities
    maxGoroutines := 4 // Adjust based on system
    if maxGoroutines > maxPaths {
        maxGoroutines = maxPaths
    }

    // Concurrency control
    sem := make(chan bool, maxGoroutines)
    
    // First, find the shortest path
    shortestPath, err := df.FindShortestPath(target)
    if err != nil {
        return nil, err
    }
    
    // Add the shortest path to results
    mutex.Lock()
    results = append(results, shortestPath)
    mutex.Unlock()
    
    // Track paths we've already found to avoid duplicates
    pathSignatures := make(map[string]bool)
    pathSignatureMutex := sync.Mutex{}
    
    // Add signature of first path
    if len(shortestPath.Path) > 0 {
        sig := getPathSignature(shortestPath.Path)
        pathSignatureMutex.Lock()
        pathSignatures[sig] = true
        pathSignatureMutex.Unlock()
    }

    // Start multiple searches with different depth limits and starting elements
    for i := 0; i < maxPaths-1; i++ {
        wg.Add(1)
        sem <- true
        
        go func(index int) {
            defer wg.Done()
            defer func() { <-sem }()
            
            // Find variation path with different depth limits and starting points
            result, err := df.findPathVariation(target, index, targetTier, pathSignatures, &pathSignatureMutex)
            
            if err == nil && result != nil {
                mutex.Lock()
                results = append(results, result)
                mutex.Unlock()
                
                // Add path signature to avoid duplicates
                sig := getPathSignature(result.Path)
                pathSignatureMutex.Lock()
                pathSignatures[sig] = true
                pathSignatureMutex.Unlock()
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
func (df *DepthFirstFinder) getPossibleRecipesThatRespectTiers(elementID string, currentTier, targetTier int) []Recipe {
    var recipes []Recipe
    
    // Get all recipes where this element is an ingredient
    for _, recipe := range df.store.Recipes {
        // Check if the current element is one of the ingredients
        isIngredient := false
        for _, ingredient := range recipe.Ingredients {
            if ingredient == elementID {
                isIngredient = true
                break
            }
        }
        
        if !isIngredient {
            continue
        }
        
        // Check if ALL ingredients have lower tier than the result
        resultTier := df.store.GetElementTier(recipe.Result)
        
        // Skip if result tier is higher than target tier (avoid going beyond what we need)
        if resultTier > targetTier {
            continue
        }
        
        allIngredientsLowerTier := true
        for _, ingredient := range recipe.Ingredients {
            ingredientTier := df.store.GetElementTier(ingredient)
            if ingredientTier >= resultTier {
                allIngredientsLowerTier = false
                break
            }
        }
        
        if allIngredientsLowerTier && resultTier > currentTier {
            recipes = append(recipes, recipe)
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

// Find variant path with different depth limits and starting points
func (df *DepthFirstFinder) findPathVariation(
    target string, 
    variationIndex int,
    targetTier int,
    existingPaths map[string]bool,
    pathMutex *sync.Mutex) (*SearchResult, error) {
    
    startTime := time.Now()

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
    
    // Rotate elements to vary the starting point
    if startElemIndex > 0 && len(reorderedElements) > 1 {
        reorderedElements = append(reorderedElements[startElemIndex:], reorderedElements[:startElemIndex]...)
    }
    
    // Try each basic element as a starting point with the new ordering
    for _, elem := range reorderedElements {
        visited[elem.ID] = true
        visitedCount++
        
        // Run DFS with depth limit and tier constraints
        found = df.dfsVariationSearch(
            elem.ID, 
            target, 
            visited, 
            parent, 
            0, 
            maxDepth, 
            targetTier, 
            &visitedCount,
            variationIndex,
            existingPaths,
            pathMutex)
        
        if found {
            // Build path and check if it's unique
            path := df.reconstructPath(target, parent)
            pathSignature := getPathSignature(path)
            
            pathMutex.Lock()
            pathExists := existingPaths[pathSignature]
            pathMutex.Unlock()
            
            // If path already exists, continue searching with the next element
            if pathExists {
                found = false
                // Reset for next element
                for k := range parent {
                    delete(parent, k)
                }
                for k := range visited {
                    delete(visited, k)
                }
                visited[elem.ID] = true
                continue
            }
            
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

// DFS variation search to find alternative paths
func (df *DepthFirstFinder) dfsVariationSearch(
    current, target string, 
    visited map[string]bool, 
    parent map[string]RecipeStep, 
    depth, maxDepth, targetTier int,
    visitedCount *int,
    variationIndex int,
    existingPaths map[string]bool,
    pathMutex *sync.Mutex) bool {
    
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
    possibleRecipes := df.getPossibleRecipesThatRespectTiers(current, currentTier, targetTier)
    
    // Add some "randomness" to the recipe order to find different paths
    if variationIndex > 0 && len(possibleRecipes) > 1 {
        // Simple shuffle based on variationIndex
        for i := 0; i < len(possibleRecipes); i++ {
            swapIndex := (i + variationIndex) % len(possibleRecipes)
            if swapIndex != i {
                possibleRecipes[i], possibleRecipes[swapIndex] = possibleRecipes[swapIndex], possibleRecipes[i]
            }
        }
    }
    
    // Try each recipe
    for _, recipe := range possibleRecipes {
        resultElem := recipe.Result
        resultTier := df.store.GetElementTier(resultElem)
        
        // Skip if any ingredient in the recipe doesn't respect tier constraint
        validRecipe := true
        for _, ingredient := range recipe.Ingredients {
            ingredientTier := df.store.GetElementTier(ingredient)
            if ingredientTier >= resultTier {
                validRecipe = false
                break
            }
        }
        
        if !validRecipe {
            continue
        }
        
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
            if df.dfsVariationSearch(resultElem, target, visited, parent, depth+1, maxDepth, targetTier, visitedCount, variationIndex, existingPaths, pathMutex) {
                return true
            }
            
            // Backtrack if needed
            delete(visited, resultElem)
            delete(parent, resultElem)
        }
    }
    
    return false
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