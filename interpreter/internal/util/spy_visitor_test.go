package util

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/domain"
)

func TestSpyVisitorCapturesResource(t *testing.T) {
	spy := NewSpyVisitor()

	resource := domain.Resource{
		Name: "host",
		Common: []domain.Field{
			{Name: "workspace_id", Required: true, Type: domain.DataType{Kind: "text"}},
		},
		Reporters: map[string][]domain.Field{
			"hbi": {
				{Name: "insights_id", Required: false, Type: domain.DataType{Kind: "uuid"}},
			},
		},
	}

	if err := spy.VisitResource(resource); err != nil {
		t.Fatalf("VisitResource failed: %v", err)
	}

	spy.AssertJSON(t, `{
		"host": {
			"common": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
			"reporters": {
				"hbi": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]
			}
		}
	}`)
}
