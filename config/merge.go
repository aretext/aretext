package config

import (
	"log"
	"reflect"
)

// MergeRecursive merges maps, slices, and simple values such as int/bool/string.
// For nil values, the non-nil value takes precedence.
// For maps of the same type, the overlay map keys are added to the base map.
// For slices of the same type, the overlay slice is appended to the base slice.
// For all other values, the overlay value replaces the base value.
func MergeRecursive(base, overlay interface{}) interface{} {
	baseType := reflect.TypeOf(base)
	overlayType := reflect.TypeOf(overlay)

	if base == nil && overlay == nil {
		return nil
	} else if base == nil {
		return overlay
	} else if overlay == nil {
		return base
	}

	if baseType.Kind() == reflect.Map && overlayType.Kind() == reflect.Map {
		return mergeMaps(base, overlay)
	}

	if baseType.Kind() == reflect.Slice && overlayType.Kind() == reflect.Slice {
		return mergeSlices(base, overlay)
	}

	return overlay
}

func mergeMaps(base, overlay interface{}) interface{} {
	baseMapValue := reflect.ValueOf(base)
	overlayMapValue := reflect.ValueOf(overlay)

	if baseMapValue.Type() != overlayMapValue.Type() {
		log.Printf("Warning: config base map type %v does not match overlay map type %v\n", baseMapValue.Type(), overlayMapValue.Type())
		return overlay
	}

	iter := overlayMapValue.MapRange()
	for iter.Next() {
		key, overlayVal := iter.Key(), iter.Value()
		baseVal := baseMapValue.MapIndex(key)
		if !baseVal.IsValid() {
			baseMapValue.SetMapIndex(key, overlayVal)
			continue
		}

		mergedVal := MergeRecursive(baseVal.Interface(), overlayVal.Interface())
		baseMapValue.SetMapIndex(key, reflect.ValueOf(mergedVal))
	}

	return base
}

func mergeSlices(base, overlay interface{}) interface{} {
	baseSliceValue := reflect.ValueOf(base)
	overlaySliceValue := reflect.ValueOf(overlay)

	if baseSliceValue.Type() != overlaySliceValue.Type() {
		log.Printf("Warning: config base slice type %v does not match overlay slice type %v\n", baseSliceValue.Type(), overlaySliceValue.Type())
		return overlay
	}

	mergedValue := reflect.AppendSlice(baseSliceValue, overlaySliceValue)
	return mergedValue.Interface()
}
