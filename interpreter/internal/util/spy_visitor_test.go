package util

import "testing"

func TestSpyVisitorCapturesResource(t *testing.T) {
	spy := NewSpyVisitor()

	spy.BeginType("host")
	common := []any{spy.VisitDataField("workspace_id", true, nil, spy.VisitTextDataType(nil, nil, nil))}
	reporters := map[string][]any{
		"hbi": {spy.VisitDataField("insights_id", false, nil, spy.VisitUUIDDataType())},
	}
	spy.VisitType("host", common, reporters)

	spy.AssertJSON(t, `{
		"host": {
			"common": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
			"reporters": {
				"hbi": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]
			}
		}
	}`)
}
