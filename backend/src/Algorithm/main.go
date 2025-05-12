package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// Element represents a game element and its recipes
type Element struct {
    ID       string     `json:"name"`
    Recipes  [][]string `json:"recipes"`
    ImageURL string     `json:"imageUrl"`
}

// ElementGroup represents a group of elements at the same tier
type ElementGroup struct {
    TierNum  int       `json:"tierNum"`
    Elements []Element `json:"elements"`
}

// Recipe represents a combination of ingredients
type Recipe struct {
    Ingredients []string
    Result      string
}

// ElementStore holds all element data
type ElementStore struct {
    Elements      map[string]*Element
    Recipes       []Recipe
    BasicElements []string
    TierMap       map[string]int // Maps element name to its tier
}

// SearchResult contains search results
type SearchResult struct {
    Path          []Recipe
    VisitedNodes  int
    ExecutionTime int64
    TreeStructure interface{}
    VariationIndex int
}

// TreeNode represents a node in the recipe tree
type TreeNode struct {
    Element    string
    Children   []*TreeNode
    IsResult   bool
    Tier       int
    ImageURL   string
}

// Common errors
var (
    ErrElementNotFound = errors.New("element not found")
    ErrNoBasicElements = errors.New("no basic elements found")
    ErrNoPathFound     = errors.New("no path found")
)

// NewElementStore creates a new element store from JSON data
func NewElementStore(jsonFile string) (*ElementStore, error) {
    file, err := os.Open(jsonFile)
    if err != nil {
        return nil, fmt.Errorf("opening element file: %w", err)
    }
    defer file.Close()

    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("reading element file: %w", err)
    }

    // Parse JSON data - now using the new structure with tiers
    var elementGroups []ElementGroup
    if err := json.Unmarshal(data, &elementGroups); err != nil {
        return nil, fmt.Errorf("parsing element JSON: %w", err)
    }

    // Create element store
    store := &ElementStore{
        Elements:      make(map[string]*Element),
        Recipes:       []Recipe{},
        BasicElements: []string{},
        TierMap:       make(map[string]int),
    }

    // Process elements
    for _, group := range elementGroups {
        for _, elem := range group.Elements {
            // Create a copy of the element for the map
            elemCopy := elem
            store.Elements[elem.ID] = &elemCopy
            
            // Track element tier
            store.TierMap[elem.ID] = group.TierNum
            
            // Identify basic elements (tier 0)
            if group.TierNum == 0 {
                store.BasicElements = append(store.BasicElements, elem.ID)
            }

            // Process recipes
            for _, ingredients := range elem.Recipes {
                if len(ingredients) == 2 {
                    recipe := Recipe{
                        Ingredients: ingredients,
                        Result:      elem.ID,
                    }
                    store.Recipes = append(store.Recipes, recipe)
                }
            }
        }
    }

    // Log summary
    log.Printf("Loaded %d elements with %d recipes", len(store.Elements), len(store.Recipes))
    log.Printf("Basic elements (tier 0): %v", store.BasicElements)

    return store, nil
}

// GetBasicElements returns basic elements
func (es *ElementStore) GetBasicElements() []*Element {
    var basics []*Element
    for _, name := range es.BasicElements {
        if elem, exists := es.Elements[name]; exists {
            basics = append(basics, elem)
        }
    }
    return basics
}

// IsBasicElement checks if element is one of the basic elements
func (es *ElementStore) IsBasicElement(elementID string) bool {
    for _, basicID := range es.BasicElements {
        if basicID == elementID {
            return true
        }
    }
    return false
}

// GetElementTier returns the tier of an element
func (es *ElementStore) GetElementTier(elementID string) int {
    tier, exists := es.TierMap[elementID]
    if !exists {
        return -1 // Unknown tier
    }
    return tier
}

// ListAvailableElements prints a sample of available elements
func ListAvailableElements(store *ElementStore, maxSample int) {
    fmt.Println("\nSome available elements:")
    
    count := 0
    for elemName := range store.Elements {
        if count < maxSample {
            tier := store.GetElementTier(elemName)
            fmt.Printf("- %s (Tier %d)\n", elemName, tier)
            count++
        } else {
            break
        }
    }
    fmt.Printf("... and %d more elements\n", len(store.Elements) - maxSample)
}

// PrintRecipePath prints the recipe path in a readable format
func PrintRecipePath(algorithmName string, result *SearchResult, store *ElementStore) {
    fmt.Printf("\n%s Path Details:\n", algorithmName)
    for i, recipe := range result.Path {
        tier1 := store.GetElementTier(recipe.Ingredients[0])
        tier2 := store.GetElementTier(recipe.Ingredients[1])
        resultTier := store.GetElementTier(recipe.Result)
        fmt.Printf("%d: %s (T%d) + %s (T%d) → %s (T%d)\n", 
            i+1, 
            recipe.Ingredients[0], tier1,
            recipe.Ingredients[1], tier2,
            recipe.Result, resultTier)
    }
}

// Recursively prints full recipe tree down to basic elements
func printFullRecipeTree(store *ElementStore, element string, recipeMap map[string][]Recipe, prefix string, isLast bool, visited map[string]bool) {
    // Prevent infinite recursion with cycles
    if visited[element] {
        fmt.Printf("%s%s %s (cycle detected)\n", prefix, getBranchChar(isLast), element)
        return
    }
    
    // Mark element as visited
    visited[element] = true
    
    // Get element tier
    tier := store.GetElementTier(element)
    
    // Format element name based on type
    var displayName string
    if store.IsBasicElement(element) {
        displayName = fmt.Sprintf("%s (T%d, BASIC)", element, tier)
    } else {
        displayName = fmt.Sprintf("%s (T%d)", element, tier)
    }
    
    // Print current element
    fmt.Printf("%s%s %s\n", prefix, getBranchChar(isLast), displayName)
    
    // If it's a basic element, stop recursion
    if store.IsBasicElement(element) {
        return
    }
    
    // Find the recipe for this element
    recipes, hasRecipe := recipeMap[element]
    if !hasRecipe || len(recipes) == 0 {
        // No recipe found, might be a basic element or missing data
        return
    }
    
    // Use the first recipe (shortest path)
    recipe := recipes[0]
    
    // Prepare child prefix
    childPrefix := prefix
    if isLast {
        childPrefix += "    "
    } else {
        childPrefix += "│   "
    }
    
    // Print the combining symbol between ingredients
    if len(recipe.Ingredients) > 1 {
        combinePrefix := childPrefix + "│"
        fmt.Printf("%s\n", combinePrefix)
        fmt.Printf("%s%s %s\n", childPrefix, "├", "Combining:")
    }
    
    // Print ingredients
    for i, ingredient := range recipe.Ingredients {
        isLastIngredient := (i == len(recipe.Ingredients) - 1)
        
        // Temporarily unmark current element to allow elements to be reused in recipes
        wasVisited := visited[element]
        delete(visited, element)
        
        printFullRecipeTree(store, ingredient, recipeMap, childPrefix, isLastIngredient, visited)
        
        // Restore current element's visited status
        visited[element] = wasVisited
    }
    
    // Remove from visited to allow reuse in other branches
    delete(visited, element)
}

// Helper function to get the appropriate branch character
func getBranchChar(isLast bool) string {
    if isLast {
        return "└──"
    }
    return "├──"
}

// BuildRecipeTree builds a tree structure from recipe path with target as root and basic elements as leaves
func BuildRecipeTree(store *ElementStore, target string, path []Recipe) *TreeNode {
    // Map to track recipes by their result
    recipesByResult := make(map[string][]Recipe)
    for _, recipe := range path {
        recipesByResult[recipe.Result] = append(recipesByResult[recipe.Result], recipe)
    }
    
    // Map to keep track of created nodes
    nodeMap := make(map[string]*TreeNode)
    
    // Create root node (target element)
    targetElem := store.Elements[target]
    root := &TreeNode{
        Element:  target,
        Children: []*TreeNode{},
        IsResult: true,
        Tier:     store.GetElementTier(target),
        ImageURL: targetElem.ImageURL,
    }
    nodeMap[target] = root
    
    // Build the tree top-down (from target to basic elements)
    var buildTree func(element string) *TreeNode
    buildTree = func(element string) *TreeNode {
        // Check if we already created this node
        if node, exists := nodeMap[element]; exists {
            return node
        }
        
        // Get element details
        elem, exists := store.Elements[element]
        if !exists {
            // Should not happen with valid data
            return nil
        }
        
        // Create new node
        node := &TreeNode{
            Element:  element,
            Children: []*TreeNode{},
            IsResult: !store.IsBasicElement(element),
            Tier:     store.GetElementTier(element),
            ImageURL: elem.ImageURL,
        }
        nodeMap[element] = node
        
        // If it's not a basic element, find the recipes that create it
        if !store.IsBasicElement(element) {
            recipes := recipesByResult[element]
            if len(recipes) > 0 {
                // Use the first recipe (shortest path)
                recipe := recipes[0]
                
                // Add ingredient nodes as children
                for _, ingredient := range recipe.Ingredients {
                    childNode := buildTree(ingredient)
                    if childNode != nil {
                        node.Children = append(node.Children, childNode)
                    }
                }
            }
        }
        
        return node
    }
    
    // Start building from the target
    buildTree(target)
    
    return root
}

// PrintRecipeTree prints a visual representation of the recipe tree
func PrintRecipeTree(store *ElementStore, target string, path []Recipe) {
    fmt.Println("\nRecipe Tree (Target → Basic Elements):")
    fmt.Println("(Basic elements are in UPPERCASE, other elements show tier in parentheses)")
    
    // Create map of recipes by result for faster lookup
    recipeMap := make(map[string][]Recipe)
    for _, recipe := range path {
        recipeMap[recipe.Result] = append(recipeMap[recipe.Result], recipe)
    }
    
    // Add all recipes from store to cover complete paths
    for _, recipe := range store.Recipes {
        if _, exists := recipeMap[recipe.Result]; !exists {
            recipeMap[recipe.Result] = append(recipeMap[recipe.Result], recipe)
        }
    }
    
    // Recursively print the tree from target to basic elements
    printFullRecipeTree(store, target, recipeMap, "", true, make(map[string]bool))
}

// printTreeNodeSimple recursively prints a tree node with simplified formatting
func printTreeNodeSimple(node *TreeNode, prefix string, isLast bool, store *ElementStore) {
    // Determine branch character
    branch := "└─"
    if !isLast {
        branch = "├─"
    }
    
    // Format node display based on type
    var nodeDisplay string
    if store.IsBasicElement(node.Element) {
        // Basic elements in UPPERCASE
        nodeDisplay = strings.ToUpper(node.Element)
    } else {
        // Other elements with tier info
        nodeDisplay = fmt.Sprintf("%s (T%d)", node.Element, node.Tier)
    }
    
    fmt.Printf("%s%s %s\n", prefix, branch, nodeDisplay)
    
    // Prepare prefix for children
    childPrefix := prefix
    if isLast {
        childPrefix += "  "
    } else {
        childPrefix += "│ "
    }
    
    // Print children
    for i, child := range node.Children {
        isLastChild := (i == len(node.Children) - 1)
        printTreeNodeSimple(child, childPrefix, isLastChild, store)
    }
}

func main() {
    // Load elements from JSON file
    fmt.Println("Loading element data...")
    dataPath := filepath.Join("..", "Scraper", "elements.json")
    store, err := NewElementStore(dataPath)
    if err != nil {
        log.Fatalf("Error loading elements: %v", err)
    }

    // Show a sample of available elements
    ListAvailableElements(store, 10)

    // Get target element from user input
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("\nEnter the target element to find (e.g., Glass, Brick, Metal): ")
    target, err := reader.ReadString('\n')
    if err != nil {
        log.Fatalf("Error reading input: %v", err)
    }
    
    // Clean up input (remove trailing newline and trim spaces)
    target = strings.TrimSpace(target)
    
    // Validate the target element exists
    if _, exists := store.Elements[target]; !exists {
        fmt.Printf("Element '%s' not found in the database!\n", target)
        os.Exit(1)
    }

    // Get algorithm choice from user
    fmt.Println("\nSelect algorithm to use:")
    fmt.Println("1. Breadth-First Search (BFS)")
    fmt.Println("2. Depth-First Search (DFS)")
    fmt.Println("3. Bidirectional Search")
    fmt.Print("Enter your choice (1-3): ")
    
    algoChoice, err := reader.ReadString('\n')
    if err != nil {
        log.Fatalf("Error reading input: %v", err)
    }
    algoChoice = strings.TrimSpace(algoChoice)

    fmt.Printf("\nSearching for recipes to create: %s (Tier %d)\n", 
        target, store.GetElementTier(target))

    // Execute the chosen algorithm
    switch algoChoice {
    case "1":
        // Run BFS search
        fmt.Println("\nRunning BFS search...")
        startTime := time.Now()
        bfs := NewBreadthFirstFinder(store)
        bfsResult, err := bfs.FindShortestPath(target)
        searchDuration := time.Since(startTime)
        
        if err != nil {
            fmt.Printf("BFS Error: %v\n", err)
        } else {
            fmt.Printf("\nBFS found a path with %d steps!\n", len(bfsResult.Path))
            fmt.Printf("Visited %d nodes during search\n", bfsResult.VisitedNodes)
            fmt.Printf("Algorithm execution time: %d ms\n", bfsResult.ExecutionTime)
            fmt.Printf("Total execution time: %v\n", searchDuration)
            
            // Print recipe path
            PrintRecipePath("BFS", bfsResult, store)
            
            // Print recipe tree
            PrintRecipeTree(store, target, bfsResult.Path)
        }
        
    case "2":
        // Run DFS search
        fmt.Println("\nRunning DFS search...")
        startTime := time.Now()
        dfs := NewDepthFirstFinder(store)
        dfsResult, err := dfs.FindShortestPath(target)
        searchDuration := time.Since(startTime)
        
        if err != nil {
            fmt.Printf("DFS Error: %v\n", err)
        } else {
            fmt.Printf("\nDFS found a path with %d steps!\n", len(dfsResult.Path))
            fmt.Printf("Visited %d nodes during search\n", dfsResult.VisitedNodes)
            fmt.Printf("Algorithm execution time: %d ms\n", dfsResult.ExecutionTime) 
            fmt.Printf("Total execution time: %v\n", searchDuration)
            
            // Print recipe path
            PrintRecipePath("DFS", dfsResult, store)
            
            // Print recipe tree
            PrintRecipeTree(store, target, dfsResult.Path)
        }
        
    case "3":
        // Run Bidirectional search
        fmt.Println("\nRunning Bidirectional search...")
        startTime := time.Now()
        bid := NewBidirectionalFinder(store)
        bidResult, err := bid.FindShortestPath(target)
        searchDuration := time.Since(startTime)
        
        if err != nil {
            fmt.Printf("Bidirectional Error: %v\n", err)
        } else {
            fmt.Printf("\nBidirectional search found a path with %d steps!\n", len(bidResult.Path))
            fmt.Printf("Visited %d nodes during search\n", bidResult.VisitedNodes)
            fmt.Printf("Algorithm execution time: %d ms\n", bidResult.ExecutionTime)
            fmt.Printf("Total execution time: %v\n", searchDuration)
            
            // Print recipe path
            PrintRecipePath("Bidirectional", bidResult, store)
            
            // Print recipe tree
            PrintRecipeTree(store, target, bidResult.Path)
        }
        
    default:
        fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
        os.Exit(1)
    }
}