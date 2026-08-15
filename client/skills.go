package client

import (
	"context"
	"encoding/json"
	"fmt"

	"go.klarlabs.de/mcp/server"
)

// Skill is one entry from skill://index.json (SEP-2640).
type Skill = server.SkillIndexEntry

// ListSkills reads skill://index.json. An absent index returns (nil, nil) so
// callers can treat enumeration as optional per the SEP.
func (c *Client) ListSkills(ctx context.Context) ([]Skill, error) {
	content, err := c.ReadResource(ctx, server.SkillIndexURI)
	if err != nil {
		// Not found / method errors surface as protocol errors from call.
		return nil, fmt.Errorf("list skills: %w", err)
	}
	if content == nil || content.Text == "" {
		return nil, nil
	}
	var idx server.SkillIndex
	if err := json.Unmarshal([]byte(content.Text), &idx); err != nil {
		return nil, fmt.Errorf("list skills: decode index: %w", err)
	}
	return idx.Skills, nil
}

// ReadSkillURI wraps resources/read for a skill:// URI (or any skill resource
// URI the server serves). Works whether or not the skill appears in the index.
func (c *Client) ReadSkillURI(ctx context.Context, uri string) (*ResourceContent, error) {
	return c.ReadResource(ctx, uri)
}
