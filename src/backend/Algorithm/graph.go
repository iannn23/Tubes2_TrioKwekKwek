package algorithm

// Element represents an element in Little Alchemy 2
type Element struct {
    Name    string     `json:"name"`
    Image   string     `json:"image"`
    Recipes [][]string `json:"recipes"`
}

// Graph represents the directed graph for Little Alchemy 2 recipes
type Graph struct {
    Forward  map[string][][]string // elem -> list of recipes ([elem1, elem2])
    Backward map[string][]string   // elem -> list of elements that produce this elem
}

// BuildGraph constructs the forward and backward graph from elements
func BuildGraph(elements []Element) Graph {
    graph := Graph{
        Forward:  make(map[string][][]string),
        Backward: make(map[string][]string),
    }

    // Build forward and backward mappings
    for _, elem := range elements {
        graph.Forward[elem.Name] = elem.Recipes
        for _, recipe := range elem.Recipes {
            if len(recipe) != 2 {
                continue
            }
            elem1, elem2 := recipe[0], recipe[1]
            // Backward mapping: elem1 and elem2 can produce elem.Name
            graph.Backward[elem1] = append(graph.Backward[elem1], elem.Name)
            graph.Backward[elem2] = append(graph.Backward[elem2], elem.Name)
        }
    }

    return graph
}