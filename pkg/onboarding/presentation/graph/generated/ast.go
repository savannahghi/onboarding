package generated

import "github.com/vektah/gqlparser/v2/ast"

// Sources exports the gglgen ast sources.
//
// These sources are used in a custom generate command to generate code using
// a "remote" schema.
//
// Each time we implement Sourceable, we need to add the new sources to the generator.
func Sources() []*ast.Source {
	return sources
}
