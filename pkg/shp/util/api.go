package util

import "errors"

// APIVerb verb name, to be called against any given resource.
type APIVerb string

const (
	// Create create api verb.
	Create APIVerb = "create"
	// Update update api verb.
	Update APIVerb = "update"
	// Delete delete api verb.
	Delete APIVerb = "delete"
)

// ErrUnknownVerb for unknown verbs, not exported as a constant.
var ErrUnknownVerb = errors.New("unknown API verb")
