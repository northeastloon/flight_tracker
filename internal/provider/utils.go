package provider

func optionalString(v any) *string {
	if s, ok := v.(string); ok {
		return &s
	}
	return nil
}

func optionalFloat(v any) *float64 {
	if f, ok := v.(float64); ok {
		return &f
	}
	return nil
}

func optionalInt64(v any) *int64 {
	if f, ok := v.(float64); ok {
		i := int64(f)
		return &i
	}
	return nil
}

func optionalIntSlice(v any) *[]int {
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	ints := make([]int, 0, len(raw))
	for _, val := range raw {
		if f, ok := val.(float64); ok {
			ints = append(ints, int(f))
		}
	}
	return &ints
}

func mustString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mustInt64(v any) int64 {
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
}

func mustBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func mustInt(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}
