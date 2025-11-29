package widgets

import (
	"fmt"
	"net/http"
	"regexp"
)

var keyRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

type Spec struct {
	Key     string
	Title   string
	Handler http.Handler

	Trigger string
	Class   string
}

type Registry struct {
	order []Spec
	byKey map[string]Spec
}

func New() *Registry {
	return &Registry{
		order: make([]Spec, 0, 8),
		byKey: make(map[string]Spec),
	}
}

func (r *Registry) Add(s Spec) error {
	if s.Handler == nil {
		return fmt.Errorf("widget %q: handler is nil", s.Key)
	}
	if !keyRe.MatchString(s.Key) {
		return fmt.Errorf("widget key %q is invalid (must match %s)", s.Key, keyRe.String())
	}
	if s.Title == "" {
		return fmt.Errorf("widget %q: title is required", s.Key)
	}
	if _, exists := r.byKey[s.Key]; exists {
		return fmt.Errorf("widget %q already registered", s.Key)
	}

	r.byKey[s.Key] = s
	r.order = append(r.order, s)
	return nil
}

func (r *Registry) MustAdd(s Spec) {
	if err := r.Add(s); err != nil {
		panic(err)
	}
}

func (r *Registry) List() []Spec {
	out := make([]Spec, len(r.order))
	copy(out, r.order)
	return out
}
