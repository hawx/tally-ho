package blog

import (
	"html/template"
	"path/filepath"
)

func ParseTemplates(webPath string) (*template.Template, error) {
	glob := filepath.Join(webPath, "template/*.gotmpl")

	return template.New("t").Funcs(template.FuncMap{
		"has":     templateHas,
		"getOr":   templateGetOr,
		"mustGet": templateMustGet,
		"content": templateContent,
	}).ParseGlob(glob)
}

func templateHas(m map[string][]interface{}, key string) bool {
	value, ok := m[key]

	return ok && len(value) > 0
}

func templateGetOr(m map[string][]interface{}, key string, or interface{}) interface{} {
	value, ok := m[key]

	if ok && len(value) > 0 {
		return value[0]
	}

	return or
}

func templateMustGet(m map[string][]interface{}, key string) interface{} {
	return m[key][0]
}

func templateContent(m map[string][]interface{}) interface{} {
	contents, ok := m["content"]

	if ok && len(contents) > 0 {
		if content, ok := contents[0].(string); ok {
			return content
		}

		if content, ok := contents[0].(map[string]interface{}); ok {
			if html, ok := content["html"]; ok {
				return template.HTML(html.(string))
			}

			if text, ok := content["text"]; ok {
				return text
			}
		}
	}

	return ""
}
