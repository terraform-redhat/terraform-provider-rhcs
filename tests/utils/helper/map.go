package helper

type m = map[string]string

// combine two strings maps to one,
// if key already exists - do nothing
func MergeMaps(map1, map2 m) m {
	for k, v := range map2 {
		_, ok := map1[k]
		if !ok {
			map1[k] = v
		}
	}
	return map1
}

// Create a file for usage
func CopyStringMap(originalMap m) m {
	newMap := make(m)
	for k, v := range originalMap {
		newMap[k] = v
	}
	return newMap
}
