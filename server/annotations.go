package server

// ToolAnnotations provides metadata hints about tool behavior.
// These help clients understand what a tool does without calling it.
type ToolAnnotations struct {
	// Title is a human-readable title for the tool.
	Title string `json:"title,omitempty"`

	// ReadOnlyHint indicates the tool only reads data (no side effects).
	// Default: false (tool might modify state)
	ReadOnlyHint *bool `json:"readOnlyHint,omitempty"`

	// DestructiveHint indicates the tool might make destructive changes.
	// Default: true (tools are assumed potentially destructive)
	DestructiveHint *bool `json:"destructiveHint,omitempty"`

	// IdempotentHint indicates calling the tool multiple times has the same
	// effect as calling it once (for the same input).
	// Default: false (tools are assumed non-idempotent)
	IdempotentHint *bool `json:"idempotentHint,omitempty"`

	// OpenWorldHint indicates the tool interacts with external systems
	// outside of the MCP host environment.
	// Default: true (tools are assumed to potentially access external systems)
	OpenWorldHint *bool `json:"openWorldHint,omitempty"`
}

// ResourceAnnotations provides metadata hints about resource behavior.
type ResourceAnnotations struct {
	// Audience describes who the resource content is intended for.
	// Values: "user" (for human consumption), "assistant" (for LLM use).
	Audience []string `json:"audience,omitempty"`

	// Priority suggests relative priority of this resource (0.0 to 1.0).
	// Higher values indicate higher priority.
	Priority *float64 `json:"priority,omitempty"`
}

// PromptAnnotations provides metadata hints about prompt behavior.
type PromptAnnotations struct {
	// Audience describes who the prompt result is intended for.
	Audience []string `json:"audience,omitempty"`

	// Priority suggests relative priority of this prompt (0.0 to 1.0).
	Priority *float64 `json:"priority,omitempty"`
}

// Bool returns a pointer to a bool value for use in annotations.
func Bool(v bool) *bool {
	return &v
}

// Float returns a pointer to a float64 value for use in annotations.
func Float(v float64) *float64 {
	return &v
}

// ReadOnly sets the tool as read-only (no side effects).
func (b *ToolBuilder) ReadOnly() *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.ReadOnlyHint = Bool(true)
	b.tool.annotations.DestructiveHint = Bool(false)
	return b
}

// Destructive marks the tool as potentially destructive.
func (b *ToolBuilder) Destructive() *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.DestructiveHint = Bool(true)
	return b
}

// Idempotent marks the tool as idempotent (multiple calls have same effect).
func (b *ToolBuilder) Idempotent() *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.IdempotentHint = Bool(true)
	return b
}

// OpenWorld marks the tool as accessing external systems.
func (b *ToolBuilder) OpenWorld() *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.OpenWorldHint = Bool(true)
	return b
}

// ClosedWorld marks the tool as not accessing external systems.
func (b *ToolBuilder) ClosedWorld() *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.OpenWorldHint = Bool(false)
	return b
}

// Title sets a human-readable title for the tool.
func (b *ToolBuilder) Title(title string) *ToolBuilder {
	if b.err != nil {
		return b
	}
	if b.tool.annotations == nil {
		b.tool.annotations = &ToolAnnotations{}
	}
	b.tool.annotations.Title = title
	return b
}

// Annotations sets custom tool annotations.
func (b *ToolBuilder) Annotations(annotations ToolAnnotations) *ToolBuilder {
	if b.err != nil {
		return b
	}
	b.tool.annotations = &annotations
	return b
}

// Audience sets the intended audience for the resource.
// Common values: "user" (human consumption), "assistant" (LLM use).
func (b *ResourceBuilder) Audience(audience ...string) *ResourceBuilder {
	if b.err != nil {
		return b
	}
	if b.resource.annotations == nil {
		b.resource.annotations = &ResourceAnnotations{}
	}
	b.resource.annotations.Audience = audience
	return b
}

// Priority sets the priority hint for the resource (0.0 to 1.0).
func (b *ResourceBuilder) Priority(priority float64) *ResourceBuilder {
	if b.err != nil {
		return b
	}
	if b.resource.annotations == nil {
		b.resource.annotations = &ResourceAnnotations{}
	}
	b.resource.annotations.Priority = Float(priority)
	return b
}

// ResourceAnnotate sets custom resource annotations.
func (b *ResourceBuilder) Annotate(annotations ResourceAnnotations) *ResourceBuilder {
	if b.err != nil {
		return b
	}
	b.resource.annotations = &annotations
	return b
}

// Audience sets the intended audience for the prompt.
func (b *PromptBuilder) Audience(audience ...string) *PromptBuilder {
	if b.err != nil {
		return b
	}
	if b.prompt.annotations == nil {
		b.prompt.annotations = &PromptAnnotations{}
	}
	b.prompt.annotations.Audience = audience
	return b
}

// Priority sets the priority hint for the prompt (0.0 to 1.0).
func (b *PromptBuilder) Priority(priority float64) *PromptBuilder {
	if b.err != nil {
		return b
	}
	if b.prompt.annotations == nil {
		b.prompt.annotations = &PromptAnnotations{}
	}
	b.prompt.annotations.Priority = Float(priority)
	return b
}

// PromptAnnotate sets custom prompt annotations.
func (b *PromptBuilder) Annotate(annotations PromptAnnotations) *PromptBuilder {
	if b.err != nil {
		return b
	}
	b.prompt.annotations = &annotations
	return b
}
