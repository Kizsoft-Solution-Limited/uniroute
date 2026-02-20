package providers

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NormalizeMessageContent converts message content from API (which may be
// []interface{} after JSON unmarshal, or a JSON string) into either a string
// or []ContentPart so providers can handle multimodal content correctly.
func NormalizeMessageContent(content interface{}) (text string, parts []ContentPart) {
	switch c := content.(type) {
	case string:
		s := strings.TrimSpace(c)
		if strings.HasPrefix(s, "[") {
			var arr []interface{}
			if err := json.Unmarshal([]byte(c), &arr); err == nil {
				_, fromArr := normalizePartsFromSlice(arr)
				if len(fromArr) > 0 {
					return "", fromArr
				}
			}
		}
		return c, nil
	case []ContentPart:
		return "", c
	case []interface{}:
		_, parts = normalizePartsFromSlice(c)
		if len(parts) > 0 {
			return "", parts
		}
		return "", nil
	default:
		return fmt.Sprintf("%v", content), nil
	}
}

func normalizePartsFromSlice(c []interface{}) (hasAny bool, parts []ContentPart) {
	for _, p := range c {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		partType, _ := pm["type"].(string)
		part := ContentPart{Type: partType}
		if t, ok := pm["text"].(string); ok {
			part.Text = t
		}
		if iu, ok := pm["image_url"].(map[string]interface{}); ok {
			if url, ok := iu["url"].(string); ok {
				part.ImageURL = &ImageURL{URL: url}
			}
		}
		if au, ok := pm["audio_url"].(map[string]interface{}); ok {
			if url, ok := au["url"].(string); ok {
				part.AudioURL = &AudioURL{URL: url}
			}
		}
		parts = append(parts, part)
		hasAny = true
	}
	return hasAny, parts
}
