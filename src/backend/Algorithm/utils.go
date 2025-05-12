package algorithm

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func containsAll(have []string, need []string) bool {
	count := map[string]int{}
	for _, h := range have {
		count[h]++
	}
	for _, n := range need {
		if count[n] == 0 {
			return false
		}
		count[n]--
	}
	return true
}
