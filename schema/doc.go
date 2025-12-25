// Package schema provides JSON Schema generation from Go types.
//
// This package automatically generates JSON Schema definitions from Go structs,
// supporting common Go types and struct tags for customization.
//
// # Basic Usage
//
// Generate a schema from a Go value:
//
//	type Person struct {
//	    Name string `json:"name" jsonschema:"required"`
//	    Age  int    `json:"age"`
//	}
//
//	schema, err := schema.Generate(Person{})
//
// # Supported Types
//
// The generator supports the following Go types:
//
//   - Structs: Converted to JSON objects with properties
//   - Strings: Converted to JSON string type
//   - Integers (all sizes): Converted to JSON integer type
//   - Floats: Converted to JSON number type
//   - Booleans: Converted to JSON boolean type
//   - Slices/Arrays: Converted to JSON array type
//   - Maps: Converted to JSON object type
//   - Pointers: Dereferenced and converted based on element type
//
// # Struct Tags
//
// The package recognizes the following struct tags:
//
//	type Example struct {
//	    // json tag controls field name
//	    Name string `json:"name"`
//
//	    // jsonschema:"required" marks field as required
//	    Required string `json:"required" jsonschema:"required"`
//
//	    // jsonschema:"description=..." adds description
//	    Desc string `json:"desc" jsonschema:"description=Field description"`
//
//	    // json:"-" excludes field
//	    Ignored string `json:"-"`
//	}
//
// # Generated Schema
//
// The Schema type represents a JSON Schema:
//
//	type Schema struct {
//	    Type        string             `json:"type,omitempty"`
//	    Properties  map[string]*Schema `json:"properties,omitempty"`
//	    Required    []string           `json:"required,omitempty"`
//	    Description string             `json:"description,omitempty"`
//	    Items       *Schema            `json:"items,omitempty"`
//	}
package schema
