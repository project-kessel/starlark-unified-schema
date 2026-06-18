package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpyVisitorCapturesResource(t *testing.T) {
	spy := NewSpyVisitor()

	spy.BeginType("host")
	common := []any{spy.VisitDataField("workspace_id", true, nil, spy.VisitTextDataType(nil, nil, nil))}
	hbiFields := []any{spy.VisitDataField("insights_id", false, nil, spy.VisitUUIDDataType())}
	err := spy.VisitResource("host", "hbi", common, hbiFields, []any{}, []any{})
	assert.NoError(t, err)

	spy.AssertJSON(t, `{
		"host": {
			"common": [{"name": "workspace_id", "required": true, "type": {"kind": "text", "minLength": null, "maxLength": null, "regex": null}}],
			"reporters": {
				"hbi": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]
			}
		}
	}`)
}
