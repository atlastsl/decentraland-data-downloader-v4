package helpers

func ArrayFilter[T any](arr []T, filter func(T) bool) (ret []T) {
	for _, ss := range arr {
		if filter(ss) {
			ret = append(ret, ss)
		}
	}
	return
}

func ArrayFilterIndex[T any](arr []T, filter func(T, int) bool) (ret []T) {
	for i, ss := range arr {
		if filter(ss, i) {
			ret = append(ret, ss)
		}
	}
	return
}

func ArrayMap[T, U any](arr []T, mapper func(T) (bool, U), skipNotExists bool, defaultValue U) (ret []U) {
	for _, ss := range arr {
		ok, u := mapper(ss)
		if ok {
			ret = append(ret, u)
		} else if !skipNotExists {
			ret = append(ret, defaultValue)
		}
	}
	return
}
