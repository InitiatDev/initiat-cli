package engine

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Registry struct {
	byID  map[string]Language
	byExt map[string][]Language
	all   []Language
}

func NewRegistry() *Registry {
	return &Registry{
		byID:  make(map[string]Language),
		byExt: make(map[string][]Language),
	}
}

func (r *Registry) Register(lang Language) error {
	if lang == nil {
		return fmt.Errorf("language cannot be nil")
	}
	id := strings.TrimSpace(lang.ID())
	if id == "" {
		return fmt.Errorf("language id cannot be empty")
	}
	if _, ok := r.byID[id]; ok {
		return fmt.Errorf("language id already registered: %q", id)
	}

	r.byID[id] = lang
	r.all = append(r.all, lang)

	for _, ext := range lang.Detector().Extensions() {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		r.byExt[ext] = append(r.byExt[ext], lang)
	}

	return nil
}

func (r *Registry) Get(id string) (Language, bool) {
	lang, ok := r.byID[id]
	return lang, ok
}

func (r *Registry) DetectByPath(path string) (Language, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	candidates := r.byExt[ext]
	if len(candidates) == 0 {
		candidates = r.all
	}

	for _, lang := range candidates {
		if lang.Detector().MatchesPath(path) {
			return lang, true
		}
	}
	return nil, false
}
