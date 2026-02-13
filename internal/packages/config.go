package packages

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"time"
)

const configXMLPath = "/conf/config.xml"

func readConfigXMLValue(path []string) (string, bool) {
	data, err := os.ReadFile(configXMLPath)
	if err != nil {
		return "", false
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	stack := make([]string, 0, len(path))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if matchesPath(stack, path) {
				val := strings.TrimSpace(string(t))
				if val != "" {
					return val, true
				}
			}
		}
	}
}

func readConfigXMLValueRetry(path []string, attempts int) (string, bool) {
	if attempts <= 0 {
		attempts = 3
	}
	for i := 0; i < attempts; i++ {
		if val, ok := readConfigXMLValue(path); ok {
			return val, true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return "", false
}

func readConfigXMLValueLoose(path []string) (string, bool) {
	data, err := os.ReadFile(configXMLPath)
	if err != nil {
		return "", false
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	stack := make([]string, 0, len(path))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if matchesPathLoose(stack, path) {
				val := strings.TrimSpace(string(t))
				if val != "" {
					return val, true
				}
			}
		}
	}
}

func readConfigXMLValueLooseRetry(path []string, attempts int) (string, bool) {
	if attempts <= 0 {
		attempts = 3
	}
	for i := 0; i < attempts; i++ {
		if val, ok := readConfigXMLValueLoose(path); ok {
			return val, true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return "", false
}

func matchesPath(stack, path []string) bool {
	if len(path) == 0 || len(stack) < len(path) {
		return false
	}
	// Strict match: contiguous suffix. This allows callers to omit the root
	// element (e.g. <pfsense> in pfSense config.xml) while still avoiding the
	// overly-permissive "loose" matcher.
	start := len(stack) - len(path)
	for i := range path {
		if stack[start+i] != path[i] {
			return false
		}
	}
	return true
}

func matchesPathLoose(stack, path []string) bool {
	if len(path) == 0 {
		return false
	}
	j := 0
	for i := 0; i < len(stack) && j < len(path); i++ {
		if stack[i] == path[j] {
			j++
		}
	}
	return j == len(path)
}

func readJSONBool(path, key string) (bool, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return false, false
	}
	v, ok := obj[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func readJSONString(path, key string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", false
	}
	v, ok := obj[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return strings.TrimSpace(s), ok
}
