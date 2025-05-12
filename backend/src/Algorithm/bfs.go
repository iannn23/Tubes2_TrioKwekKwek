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
    targetTier := bf.store.GetElementTier(target)

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
        possibleRecipes := bf.getPossibleRecipesThatRespectTiers(current, currentTier, targetTier)
        for _, recipe := range possibleRecipes {
            resultElem := recipe.Result
            resultTier := bf.store.GetElementTier(resultElem)
            
            // Skip if any ingredient in the recipe doesn't respect tier constraint
            validRecipe := true
            for _, ingredient := range recipe.Ingredients {
                ingredientTier := bf.store.GetElementTier(ingredient)
                if ingredientTier >= resultTier {
                    validRecipe = false
                    break
                }
            }
            
            if !validRecipe {
                continue
            }
            
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

// FindMultiplePaths finds multiple recipe paths using multithreading
func (bf *BreadthFirstFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target exists
    _, exists := bf.store.Elements[target]  // Remove targetElem
    if !exists {
        return nil, ErrElementNotFound
    }
    targetTier := bf.store.GetElementTier(target)

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
    shortestPath, err := bf.FindShortestPath(target)
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

    // Start multiple searches with different variations
    for i := 0; i < maxPaths-1; i++ {
        wg.Add(1)
        sem <- true

        go func(index int) {
            defer wg.Done()
            defer func() { <-sem }()

            // Find variation path with constraints on tier and ingredients
            result, err := bf.findPathVariation(target, index, targetTier, pathSignatures, &pathSignatureMutex)

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

// Get a unique signature for a path to detect duplicates
func getPathSignature(path []Recipe) string {
    if len(path) == 0 {
        return ""
    }
    
    // Simple signature based on sequence of results
    sig := ""
    for _, recipe := range path {
        sig += recipe.Result + ":"
        // Sort ingredients to normalize the recipe
        if len(recipe.Ingredients) >= 2 {
            if recipe.Ingredients[0] > recipe.Ingredients[1] {
                sig += recipe.Ingredients[1] + "+" + recipe.Ingredients[0] + ";"
            } else {
                sig += recipe.Ingredients[0] + "+" + recipe.Ingredients[1] + ";"
            }
        }
    }
    return sig
}

// Get recipes using element that respect tier hierarchy
func (bf *BreadthFirstFinder) getPossibleRecipesThatRespectTiers(elementID string, currentTier, targetTier int) []Recipe {
    var recipes []Recipe
    
    // Get first all recipes where this element is an ingredient
    for _, recipe := range bf.store.Recipes {
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
        resultTier := bf.store.GetElementTier(recipe.Result)
        
        // Skip if result tier is higher than target tier (avoid going beyond what we need)
        if resultTier > targetTier {
            continue
        }
        
        allIngredientsLowerTier := true
        for _, ingredient := range recipe.Ingredients {
            ingredientTier := bf.store.GetElementTier(ingredient)
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

// Find a path variation using BFS with constraints
func (bf *BreadthFirstFinder) findPathVariation(
    target string, 
    variationIndex int, 
    targetTier int,
    existingPaths map[string]bool, 
    pathMutex *sync.Mutex) (*SearchResult, error) {
    
    startTime := time.Now()

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

    // Initialize queue with basic elements, but in a different order based on variationIndex
    // This helps us find different paths
    offset := variationIndex % len(basicElements)
    for i := 0; i < len(basicElements); i++ {
        elemIndex := (i + offset) % len(basicElements)
        elem := basicElements[elemIndex]
        queue.PushBack(elem.ID)
        visited[elem.ID] = true
        visitedCount++
    }

    // Path found flag
    found := false

    // Add some "randomness" in recipe selection based on variationIndex
    randomFactor := variationIndex % 10
    
    // Run BFS with some variations
    for queue.Len() > 0 && !found {
        current := queue.Remove(queue.Front()).(string)
        currentTier := bf.store.GetElementTier(current)

        // Check if we found the target
        if current == target {
            found = true
            break
        }

        // Get possible recipes
        possibleRecipes := bf.getPossibleRecipesThatRespectTiers(current, currentTier, targetTier)
        
        // Reorder recipes based on variationIndex to encourage finding different paths
        if len(possibleRecipes) > 1 && randomFactor > 0 {
            // Simple reordering: swap elements based on variationIndex
            for i := 0; i < len(possibleRecipes)-1; i += 2 {
                if i+1 < len(possibleRecipes) && (i+randomFactor)%7 == 0 {
                    possibleRecipes[i], possibleRecipes[i+1] = possibleRecipes[i+1], possibleRecipes[i]
                }
            }
        }
        
        for _, recipe := range possibleRecipes {
            resultElem := recipe.Result
            resultTier := bf.store.GetElementTier(resultElem)
            
            // Skip if any ingredient in the recipe doesn't respect tier constraint
            validRecipe := true
            for _, ingredient := range recipe.Ingredients {
                ingredientTier := bf.store.GetElementTier(ingredient)
                if ingredientTier >= resultTier {
                    validRecipe = false
                    break
                }
            }
            
            if !validRecipe {
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
                    
                    // Check if this path is unique compared to existing ones
                    path := bf.reconstructPath(target, parent)
                    pathSignature := getPathSignature(path)
                    
                    pathMutex.Lock()
                    pathExists := existingPaths[pathSignature]
                    pathMutex.Unlock()
                    
                    if pathExists {
                        // This path is a duplicate, continue searching
                        found = false
                    }
                    
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
        VariationIndex: variationIndex,
    }, nil
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