package main

import (
    "container/list"
    "fmt"
    "sync"
    "time"
)

// BidirectionalFinder for recipe search
type BidirectionalFinder struct {
    store *ElementStore
}

// NewBidirectionalFinder creates finder instance
func NewBidirectionalFinder(store *ElementStore) *BidirectionalFinder {
    return &BidirectionalFinder{store: store}
}

// FindShortestPath finds shortest recipe path
func (bf *BidirectionalFinder) FindShortestPath(target string) (*SearchResult, error) {
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

    // Forward search setup
    forwardQueue := list.New()
    forwardVisited := make(map[string]bool)
    forwardParent := make(map[string]RecipeStep) // Parent tracking
    forwardTier := make(map[string]int)         // Track tier for each element in forward search

    // Init forward queue
    for _, elem := range basicElements {
        forwardQueue.PushBack(elem.ID)
        forwardVisited[elem.ID] = true
        forwardTier[elem.ID] = 0 // Basic elements are tier 0
        visitedCount++
    }

    // Backward search setup
    backwardQueue := list.New()
    backwardVisited := make(map[string]bool)
    backwardParent := make(map[string]RecipeStep) // Child tracking
    backwardTier := make(map[string]int)         // Track tier for each element in backward search

    // Init backward queue
    backwardQueue.PushBack(target)
    backwardVisited[target] = true
    backwardTier[target] = targetTier
    visitedCount++

    // Meeting point
    var meetingPoint string
    found := false

    // Run bidirectional BFS with tier constraints
    for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 && !found {
        // Forward search step
        levelSize := forwardQueue.Len()
        for i := 0; i < levelSize && !found; i++ {
            current := forwardQueue.Remove(forwardQueue.Front()).(string)
            currentTier := forwardTier[current]
            
            // Check meeting point
            if backwardVisited[current] {
                meetingPoint = current
                found = true
                break
            }

            // Expand forward - respect tier hierarchy
            possibleRecipes := bf.getValidRecipesWithElementAnyPosition(current)
            for _, recipe := range possibleRecipes {
                resultElem := recipe.Result
                resultTier := bf.store.GetElementTier(resultElem)
                
                // Check if ALL ingredients have lower tier than the result
                allIngredientsLowerTier := true
                for _, ingredient := range recipe.Ingredients {
                    ingredientTier := bf.store.GetElementTier(ingredient)
                    if ingredientTier >= resultTier {
                        allIngredientsLowerTier = false
                        break
                    }
                }
                
                // Only process recipes where all ingredients have lower tiers than result
                if !allIngredientsLowerTier {
                    continue
                }
                
                // Only consider recipes that lead to higher tiers
                if resultTier <= currentTier {
                    continue
                }
                
                if !forwardVisited[resultElem] {
                    forwardQueue.PushBack(resultElem)
                    forwardVisited[resultElem] = true
                    forwardTier[resultElem] = resultTier
                    forwardParent[resultElem] = RecipeStep{
                        ParentID: current,
                        Recipe:   recipe,
                    }
                    visitedCount++
                    
                    // Check if we've met the backward search
                    if backwardVisited[resultElem] {
                        meetingPoint = resultElem
                        found = true
                        break
                    }
                }
            }
        }

        if found {
            break
        }

        // Backward search step
        levelSize = backwardQueue.Len()
        for i := 0; i < levelSize && !found; i++ {
            current := backwardQueue.Remove(backwardQueue.Front()).(string)
            currentTier := backwardTier[current]
            
            // Check meeting point
            if forwardVisited[current] {
                meetingPoint = current
                found = true
                break
            }

            // Expand backward - but respect tier hierarchy
            // In backward search, we're looking for ingredients of current element
            for _, recipe := range bf.store.Recipes {
                if recipe.Result == current {
                    // Check if all ingredients have lower tier than result (current)
                    allIngredientsLowerTier := true
                    for _, ingredient := range recipe.Ingredients {
                        ingredientTier := bf.store.GetElementTier(ingredient)
                        if ingredientTier >= currentTier {
                            allIngredientsLowerTier = false
                            break
                        }
                    }
                    
                    // Skip recipes that don't satisfy tier constraints
                    if !allIngredientsLowerTier {
                        continue
                    }
                    
                    // For each ingredient, process if it meets tier constraints
                    for _, ingredient := range recipe.Ingredients {
                        ingredientTier := bf.store.GetElementTier(ingredient)
                        
                        // Only consider ingredients from lower tiers
                        if ingredientTier >= currentTier {
                            continue
                        }
                        
                        if !backwardVisited[ingredient] {
                            backwardQueue.PushBack(ingredient)
                            backwardVisited[ingredient] = true
                            backwardTier[ingredient] = ingredientTier
                            backwardParent[ingredient] = RecipeStep{
                                ParentID: current, 
                                Recipe:   recipe,
                            }
                            visitedCount++
                            
                            // Check if we've met the forward search
                            if forwardVisited[ingredient] {
                                meetingPoint = ingredient
                                found = true
                                break
                            }
                        }
                    }
                    
                    if found {
                        break
                    }
                }
            }
        }
    }

    if !found {
        return nil, ErrNoPathFound
    }

    // Build path
    forwardPath := bf.reconstructForwardPath(meetingPoint, forwardParent)
    backwardPath := bf.reconstructBackwardPath(meetingPoint, backwardParent)

    // Combine paths
    var completePath []Recipe
    completePath = append(completePath, forwardPath...)
    completePath = append(completePath, backwardPath...)

    // Visualize tree
    treeStructure := bf.buildTreeStructure(completePath, target)

    executionTime := time.Since(startTime).Milliseconds()

    return &SearchResult{
        Path:          completePath,
        VisitedNodes:  visitedCount,
        ExecutionTime: executionTime,
        TreeStructure: treeStructure,
    }, nil
}

// FindMultiplePaths finds multiple recipe paths using multithreading
func (bf *BidirectionalFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target exists
    _, exists := bf.store.Elements[target]
    if !exists {
        return nil, ErrElementNotFound
    }

    var results []*SearchResult
    var mutex sync.Mutex
    var wg sync.WaitGroup

    // Limit goroutines based on available CPUs
    maxGoroutines := 4 // Can be adjusted based on system cores
    if maxGoroutines > maxPaths {
        maxGoroutines = maxPaths
    }

    // Channel for concurrency control
    sem := make(chan bool, maxGoroutines)
    
    // Channel to collect found paths
    foundPaths := make(chan *SearchResult, maxPaths)
    
    // Stop signal for goroutines when we've found enough paths
    done := make(chan bool)
    pathsFound := 0
    
    // Start multiple searches with different variations
    for i := 0; i < maxPaths*2; i++ { // Try more variations than needed
        wg.Add(1)
        sem <- true
        
        go func(index int) {
            defer wg.Done()
            defer func() { <-sem }()
            
            // Create variation of finder parameters based on index
            select {
            case <-done:
                return // Stop if we've found enough paths
            default:
                // Find variation path with different criteria
                result, err := bf.findPathWithVariation(target, index)
                
                if err == nil && result != nil {
                    // Verify the path is valid (all ingredients have lower tier than results)
                    isValid := true
                    for _, recipe := range result.Path {
                        resultTier := bf.store.GetElementTier(recipe.Result)
                        
                        for _, ingredient := range recipe.Ingredients {
                            ingredientTier := bf.store.GetElementTier(ingredient)
                            if ingredientTier >= resultTier {
                                isValid = false
                                break
                            }
                        }
                        
                        if !isValid {
                            break
                        }
                    }
                    
                    if isValid {
                        // Check if we already have this path (deduplicate)
                        isDuplicate := false
                        
                        mutex.Lock()
                        for _, existingResult := range results {
                            if len(existingResult.Path) == len(result.Path) {
                                // Simple equality check - could be more sophisticated
                                sameSteps := true
                                for i, recipe := range existingResult.Path {
                                    if recipe.Result != result.Path[i].Result {
                                        sameSteps = false
                                        break
                                    }
                                }
                                
                                if sameSteps {
                                    isDuplicate = true
                                    break
                                }
                            }
                        }
                        mutex.Unlock()
                        
                        if !isDuplicate {
                            foundPaths <- result
                        }
                    }
                }
            }
        }(i)
    }
    
    // Collect results in a separate goroutine
    go func() {
        for result := range foundPaths {
            mutex.Lock()
            if pathsFound < maxPaths {
                results = append(results, result)
                pathsFound++
                
                if pathsFound >= maxPaths {
                    close(done) // Signal other goroutines to stop
                }
            }
            mutex.Unlock()
        }
    }()
    
    // Wait for all searches to complete
    wg.Wait()
    close(foundPaths)
    
    // Check results
    if len(results) == 0 {
        return nil, ErrNoPathFound
    }
    
    return results, nil
}

// Get valid recipes that contain the given element in any position
// and respect tier constraints
func (bf *BidirectionalFinder) getValidRecipesWithElementAnyPosition(elementID string) []Recipe {
    var recipes []Recipe

    // Find recipes where the element is used as an ingredient
    for _, recipe := range bf.store.Recipes {
        // Check if this element is one of the ingredients
        isIngredient := false
        for _, ingredient := range recipe.Ingredients {
            if ingredient == elementID {
                isIngredient = true
                break
            }
        }
        
        if isIngredient {
            // Verify tier constraints - all ingredients must have lower tier than result
            resultTier := bf.store.GetElementTier(recipe.Result)
            allIngredientsLowerTier := true
            
            for _, ingredient := range recipe.Ingredients {
                ingredientTier := bf.store.GetElementTier(ingredient)
                if ingredientTier >= resultTier {
                    allIngredientsLowerTier = false
                    break
                }
            }
            
            if allIngredientsLowerTier {
                recipes = append(recipes, recipe)
            }
        }
    }

    return recipes
}

// Build forward path
func (bf *BidirectionalFinder) reconstructForwardPath(meetingPoint string, parentMap map[string]RecipeStep) []Recipe {
    var path []Recipe
    current := meetingPoint
    
    for {
        step, exists := parentMap[current]
        if !exists {
            break // Reached basic element
        }
        
        // Verify tier constraints for the recipe
        resultTier := bf.store.GetElementTier(step.Recipe.Result)
        allIngredientsLowerTier := true
        
        for _, ingredient := range step.Recipe.Ingredients {
            ingredientTier := bf.store.GetElementTier(ingredient)
            if ingredientTier >= resultTier {
                allIngredientsLowerTier = false
                break
            }
        }
        
        if !allIngredientsLowerTier {
            break // Skip invalid recipe
        }
        
        path = append([]Recipe{step.Recipe}, path...) // Prepend
        current = step.ParentID
    }
    
    return path
}

// Build backward path
func (bf *BidirectionalFinder) reconstructBackwardPath(meetingPoint string, childMap map[string]RecipeStep) []Recipe {
    var path []Recipe
    current := meetingPoint
    
    for {
        step, exists := childMap[current]
        if !exists {
            break
        }
        
        // Verify tier constraints for the recipe
        resultTier := bf.store.GetElementTier(step.Recipe.Result)
        allIngredientsLowerTier := true
        
        for _, ingredient := range step.Recipe.Ingredients {
            ingredientTier := bf.store.GetElementTier(ingredient)
            if ingredientTier >= resultTier {
                allIngredientsLowerTier = false
                break
            }
        }
        
        if !allIngredientsLowerTier {
            break // Skip invalid recipe
        }
        
        path = append(path, step.Recipe) // Append
        current = step.ParentID
    }
    
    return path
}

// Find variant path with different search parameters
func (bf *BidirectionalFinder) findPathWithVariation(target string, variationIndex int) (*SearchResult, error) {
    startTime := time.Now()

    // Modify search parameters based on variation index
    // This creates different search patterns to find varied paths
    
    // Different variations:
    // - Use different subsets of basic elements as starting points
    // - Modify tier prioritization
    // - Use different randomization seeds for recipe order
    
    // For simplicity, we'll implement a basic variation using
    // different starting elements and search bias
    
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

    // Forward search setup with variation
    forwardQueue := list.New()
    forwardVisited := make(map[string]bool)
    forwardParent := make(map[string]RecipeStep) // Parent tracking
    forwardTier := make(map[string]int)         // Track tier for each element in forward search

    // Use variation index to select different starting elements
    startOffset := variationIndex % len(basicElements)
    elemCount := len(basicElements)
    
    // Add basic elements in different order based on variation
    for i := 0; i < elemCount; i++ {
        elemIndex := (startOffset + i) % elemCount
        elem := basicElements[elemIndex]
        
        forwardQueue.PushBack(elem.ID)
        forwardVisited[elem.ID] = true
        forwardTier[elem.ID] = 0
        visitedCount++
    }

    // Backward search setup
    backwardQueue := list.New()
    backwardVisited := make(map[string]bool)
    backwardParent := make(map[string]RecipeStep) // Child tracking
    backwardTier := make(map[string]int)         // Track tier for each element in backward search

    // Init backward queue
    backwardQueue.PushBack(target)
    backwardVisited[target] = true
    backwardTier[target] = targetTier
    visitedCount++

    // Meeting point
    var meetingPoint string
    found := false
    
    // Set variation-specific depth limit to avoid too deep searches
    maxIterations := 1000 + (variationIndex % 500) // Vary max iterations
    iterations := 0

    // Run bidirectional BFS with tier constraints
    for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 && !found && iterations < maxIterations {
        iterations++
        
        // Forward search step with extra randomization based on variation index
        doBackwardFirst := (variationIndex % 2 == 1) // Alternate search direction
        
        if !doBackwardFirst {
            // Forward search
            levelSize := forwardQueue.Len()
            for i := 0; i < levelSize && !found; i++ {
                current := forwardQueue.Remove(forwardQueue.Front()).(string)
                currentTier := forwardTier[current]
                
                // Check meeting point
                if backwardVisited[current] {
                    meetingPoint = current
                    found = true
                    break
                }
                
                // Expand forward with tier constraints
                possibleRecipes := bf.getValidRecipesWithElementAnyPosition(current)
                
                // Process recipes in different orders based on variation
                offset := variationIndex % len(possibleRecipes)
                if offset > 0 && len(possibleRecipes) > 1 {
                    possibleRecipes = append(possibleRecipes[offset:], possibleRecipes[:offset]...)
                }
                
                for _, recipe := range possibleRecipes {
                    resultElem := recipe.Result
                    resultTier := bf.store.GetElementTier(resultElem)
                    
                    // Verify tier constraints
                    allIngredientsLowerTier := true
                    for _, ingredient := range recipe.Ingredients {
                        ingredientTier := bf.store.GetElementTier(ingredient)
                        if ingredientTier >= resultTier {
                            allIngredientsLowerTier = false
                            break
                        }
                    }
                    
                    if !allIngredientsLowerTier || resultTier <= currentTier {
                        continue
                    }
                    
                    if !forwardVisited[resultElem] {
                        forwardQueue.PushBack(resultElem)
                        forwardVisited[resultElem] = true
                        forwardTier[resultElem] = resultTier
                        forwardParent[resultElem] = RecipeStep{
                            ParentID: current,
                            Recipe:   recipe,
                        }
                        visitedCount++
                        
                        if backwardVisited[resultElem] {
                            meetingPoint = resultElem
                            found = true
                            break
                        }
                    }
                }
            }
        }
        
        if found {
            break
        }
        
        // Backward search step
        levelSize := backwardQueue.Len()
        for i := 0; i < levelSize && !found; i++ {
            current := backwardQueue.Remove(backwardQueue.Front()).(string)
            currentTier := backwardTier[current]
            
            if forwardVisited[current] {
                meetingPoint = current
                found = true
                break
            }
            
            // Process recipes in different order based on variation
            relevantRecipes := []Recipe{}
            for _, recipe := range bf.store.Recipes {
                if recipe.Result == current {
                    relevantRecipes = append(relevantRecipes, recipe)
                }
            }
            
            // Vary recipe order
            offset := variationIndex % len(relevantRecipes)
            if offset > 0 && len(relevantRecipes) > 1 {
                relevantRecipes = append(relevantRecipes[offset:], relevantRecipes[:offset]...)
            }
            
            for _, recipe := range relevantRecipes {
                // Check tier constraints
                allIngredientsLowerTier := true
                for _, ingredient := range recipe.Ingredients {
                    ingredientTier := bf.store.GetElementTier(ingredient)
                    if ingredientTier >= currentTier {
                        allIngredientsLowerTier = false
                        break
                    }
                }
                
                if !allIngredientsLowerTier {
                    continue
                }
                
                // Process ingredients differently based on variation
                processOrder := recipe.Ingredients
                if variationIndex%2 == 1 && len(processOrder) > 1 {
                    // Reverse ingredient processing order in odd variations
                    processOrder = []string{processOrder[1], processOrder[0]}
                }
                
                for _, ingredient := range processOrder {
                    ingredientTier := bf.store.GetElementTier(ingredient)
                    
                    if ingredientTier >= currentTier {
                        continue
                    }
                    
                    if !backwardVisited[ingredient] {
                        backwardQueue.PushBack(ingredient)
                        backwardVisited[ingredient] = true
                        backwardTier[ingredient] = ingredientTier
                        backwardParent[ingredient] = RecipeStep{
                            ParentID: current,
                            Recipe:   recipe,
                        }
                        visitedCount++
                        
                        if forwardVisited[ingredient] {
                            meetingPoint = ingredient
                            found = true
                            break
                        }
                    }
                }
                
                if found {
                    break
                }
            }
        }
        
        // If we did forward search first, now do backward
        if doBackwardFirst {
            // Forward search (same code as above)
            levelSize := forwardQueue.Len()
            for i := 0; i < levelSize && !found; i++ {
                current := forwardQueue.Remove(forwardQueue.Front()).(string)
                currentTier := forwardTier[current]
                
                // Check meeting point
                if backwardVisited[current] {
                    meetingPoint = current
                    found = true
                    break
                }
                
                // Expand forward with tier constraints
                possibleRecipes := bf.getValidRecipesWithElementAnyPosition(current)
                
                // Process recipes in different orders based on variation
                offset := variationIndex % len(possibleRecipes)
                if offset > 0 && len(possibleRecipes) > 1 {
                    possibleRecipes = append(possibleRecipes[offset:], possibleRecipes[:offset]...)
                }
                
                for _, recipe := range possibleRecipes {
                    resultElem := recipe.Result
                    resultTier := bf.store.GetElementTier(resultElem)
                    
                    // Verify tier constraints
                    allIngredientsLowerTier := true
                    for _, ingredient := range recipe.Ingredients {
                        ingredientTier := bf.store.GetElementTier(ingredient)
                        if ingredientTier >= resultTier {
                            allIngredientsLowerTier = false
                            break
                        }
                    }
                    
                    if !allIngredientsLowerTier || resultTier <= currentTier {
                        continue
                    }
                    
                    if !forwardVisited[resultElem] {
                        forwardQueue.PushBack(resultElem)
                        forwardVisited[resultElem] = true
                        forwardTier[resultElem] = resultTier
                        forwardParent[resultElem] = RecipeStep{
                            ParentID: current,
                            Recipe:   recipe,
                        }
                        visitedCount++
                        
                        if backwardVisited[resultElem] {
                            meetingPoint = resultElem
                            found = true
                            break
                        }
                    }
                }
            }
        }
    }

    if !found {
        return nil, ErrNoPathFound
    }

    // Build path
    forwardPath := bf.reconstructForwardPath(meetingPoint, forwardParent)
    backwardPath := bf.reconstructBackwardPath(meetingPoint, backwardParent)

    // Combine paths
    var completePath []Recipe
    completePath = append(completePath, forwardPath...)
    completePath = append(completePath, backwardPath...)

    // Visualize tree
    treeStructure := bf.buildTreeStructure(completePath, target)

    executionTime := time.Since(startTime).Milliseconds()

    return &SearchResult{
        Path:          completePath,
        VisitedNodes:  visitedCount,
        ExecutionTime: executionTime,
        TreeStructure: treeStructure,
        VariationIndex: variationIndex,
    }, nil
}

// Create visualization tree
func (bf *BidirectionalFinder) buildTreeStructure(path []Recipe, target string) interface{} {
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