package algorithm

import (
	"sync"
	"tubes2/model"
)

func FindMultipleRecipes(target string, elements map[string]*model.Element, max int) [][]string {
	var results [][]string
	var wg sync.WaitGroup
	var mu sync.Mutex

	queue := make(chan []string, 100)
	base := []string{"air", "earth", "fire", "water"}

	for _, b := range base {
		queue <- []string{b}
	}

	worker := func() {
		defer wg.Done()
		for path := range queue {
			last := path[len(path)-1]
			if last == target {
				mu.Lock()
				results = append(results, path)
				if len(results) >= max {
					mu.Unlock()
					return
				}
				mu.Unlock()
				continue
			}
			for name, el := range elements {
				if contains(path, name) {
					continue
				}
				for _, recipe := range el.Recipes {
					if containsAll(path, recipe) {
						newPath := append([]string{}, path...)
						newPath = append(newPath, name)
						queue <- newPath
					}
				}
			}
		}
	}

	n := 8
	wg.Add(n)
	for i := 0; i < n; i++ {
		go worker()
	}

	wg.Wait()
	close(queue)

	return results
}
