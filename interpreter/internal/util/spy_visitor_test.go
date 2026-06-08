package util

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/model"
)

func TestSpyVisitorCapturesResource(t *testing.T) {
	spy := NewSpyVisitor()

	resource := model.Resource{
		Name: "host",
		Common: []model.Field{
			{Name: "workspace_id", Required: true, Type: model.DataType{Kind: "text"}},
		},
		Reporters: map[string][]model.Field{
			"hbi": {
				{Name: "insights_id", Required: false, Type: model.DataType{Kind: "uuid"}},
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
