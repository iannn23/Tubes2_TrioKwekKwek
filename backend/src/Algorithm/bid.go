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
            possibleRecipes := bf.getPossibleRecipesThatRespectTiers(current, currentTier)
            for _, recipe := range possibleRecipes {
                resultElem := recipe.Result
                resultTier := bf.store.GetElementTier(resultElem)
                
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
                    // For each ingredient, check tier constraints
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

// FindMultiplePaths finds multiple recipe paths
func (bf *BidirectionalFinder) FindMultiplePaths(target string, maxPaths int) ([]*SearchResult, error) {
    // Check target
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
func (bf *BidirectionalFinder) getPossibleRecipesThatRespectTiers(elementID string, currentTier int) []Recipe {
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
func (bf *BidirectionalFinder) getPossibleRecipesWith(elementID string) []Recipe {
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

// Build forward path
func (bf *BidirectionalFinder) reconstructForwardPath(meetingPoint string, parentMap map[string]RecipeStep) []Recipe {
    var path []Recipe
    current := meetingPoint
    
    for {
        step, exists := parentMap[current]
        if !exists {
            break // Reached basic element
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
        
        path = append(path, step.Recipe) // Append
        current = step.ParentID
    }
    
    return path
}

// Find variant path
func (bf *BidirectionalFinder) findPathWithVariation(target string, variationIndex int) (*SearchResult, error) {
    result, err := bf.FindShortestPath(target)
    
    if err == nil && result != nil {
        // Add variation marker
        result.VariationIndex = variationIndex
    }
    
    return result, err
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