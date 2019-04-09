package blog

import "html/template"

func parseTemplates(glob string) (*template.Template, error) {
	return template.New("t").Funcs(template.FuncMap{
		"has": func(m map[string][]interface{}, key string) bool {
			value, ok := m[key]

			return ok && len(value) > 0
		},
		"getOr": func(m map[string][]interface{}, key string, or interface{}) interface{} {
			value, ok := m[key]

			if ok && len(value) > 0 {
				return value[0]
			}

			return or
		},
		"mustGet": func(m map[string][]interface{}, key string) interface{} {
			return m[key][0]
		},
		"content": func(m map[string][]interface{}) interface{} {
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
		},
	}).ParseGlob(glob)
}
