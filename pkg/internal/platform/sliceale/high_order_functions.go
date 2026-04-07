package sliceale

import (
	"fmt"
)

// ApplyMapFunction applies the function to each element of the slice and returns a new slice of the results in the same order.
// Don't confuse with the data structure `map`
// More info at: https://en.wikipedia.org/wiki/Map_(higher-order_function)
func ApplyMapFunction[T1, T2 any](s []T2, f func(T2) T1) []T1 {
	res := make([]T1, 0, len(s))
	for _, e := range s {
		fe := f(e)
		res = append(res, fe)
	}

	return res
}

// HasDuplicateField returns true if the slice has at least two elements with the same value when projected by the given function.
func HasDuplicateField[T1 any, T2 comparable](elems []T1, projector func(T1) T2) bool {
	accum := make(map[T2]struct{})
	for _, elem := range elems {
		if _, ok := accum[projector(elem)]; ok {
			return true
		}
		accum[projector(elem)] = struct{}{}
	}
	return false
}

// Select applies the given function to each element in the slice and returns a new slice containing only the elements
// that satisfy the function.
func Select[T any](slice []T, condition func(T) bool) []T {
	var result []T

	for _, element := range slice {
		if condition(element) {
			result = append(result, element)
		}
	}

	return result
}

// ApplyReduceFunction applies a function to each element of the slice and returns a single value.
// It can also be called a `Fold`. This specific implementation is equivalent to a foldl
// More info at: https://en.wikipedia.org/wiki/Fold_(higher-order_function)
func ApplyReduceFunction[T1, T2 any](s []T1, f func(T1, T2) T2, initVal T2) T2 {
	if s == nil {
		return initVal
	}
	accum := initVal
	for _, v := range s {
		accum = f(v, accum)
	}
	return accum
}

// GroupBy groups the elements of the slice into a new map, where the keys are the result of the projection
// and the values are the elements of the slice with that projection.
// WARNING: don't assume any order in the slices of the values of the map. The original order is not preserved.
func GroupBy[T1 any, T2 comparable](slice []T1, projection func(T1) T2) map[T2][]T1 {
	joinFn := func(elem T1, elems []T1) []T1 {
		return append(elems, elem)
	}
	return GroupByWithJoinFunction(slice, projection, joinFn)

}

// GroupByUniqueProjection groups the elements of the slice by the result of the given function. If two elements
// have the same projection, it returns an error.
func GroupByUniqueProjection[T1 any, T2 comparable](slice []T1, projection func(T1) T2) (map[T2]T1, error) {
	if slice == nil {
		return nil, nil
	}

	accum := make(map[T2]T1)
	for i, elem := range slice {
		if _, ok := accum[projection(elem)]; ok {
			return nil, fmt.Errorf(
				"duplicate projection for projection(s[i])=projection(s[%d])=projection(%v)=%v",
				i,
				elem,
				projection(elem),
			)
		}

		accum[projection(elem)] = elem
	}

	return accum, nil
}

// GroupByWithJoinFunction groups the elements of the slice by the result of the given "projection" function and joins the
// result with the previous elements in that group using the "join" function"
// WARNING: don't assume any order in the join function. The original order is not guaranteed
func GroupByWithJoinFunction[TOrig any, TKey comparable, TValue any](slice []TOrig, projection func(TOrig) TKey, join func(TOrig, TValue) TValue) map[TKey]TValue {
	if slice == nil {
		return nil
	}

	accum := make(map[TKey]TValue)
	for _, elem := range slice {
		key := projection(elem)
		if _, ok := accum[key]; ok {
			accum[key] = join(elem, accum[key])
		} else {
			accumZeroValue := *new(TValue)
			accum[key] = join(elem, accumZeroValue)
		}
	}

	return accum
}

// ApplyMapFunctionRemovingDuplicates applies the function to each element of the slice and returns a new slice of the
// results in the same order, but without duplicates.
func ApplyMapFunctionRemovingDuplicates[T1 comparable, T2 any](s []T2, f func(T2) T1) []T1 {
	if len(s) == 0 {
		return []T1{}
	}

	seen := make(map[T1]struct{})
	res := make([]T1, 0, len(s))
	for _, key := range s {
		e := f(key)
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		res = append(res, e)
	}

	return res
}

// RemoveDuplicates returns a new slice with the same elements as the input slice, but without duplicates.
func RemoveDuplicates[T1 comparable](s []T1) []T1 {
	seen := make(map[T1]struct{})
	res := make([]T1, 0, len(s))
	for _, key := range s {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		res = append(res, key)
	}

	return res

}
