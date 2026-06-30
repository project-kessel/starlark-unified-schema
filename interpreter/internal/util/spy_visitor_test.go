package util

import (
	"testing"

	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/stretchr/testify/assert"
)

func TestSpyVisitorCapturesResource(t *testing.T) {
	spy := NewSpyVisitor()

	spy.BeginType("host")
	commonMembers := &output.Members{DataFields: []any{spy.VisitDataField("workspace_id", true, nil, spy.VisitTextDataType(nil, nil, nil))}}
	reporterMembers := &output.Members{DataFields: []any{spy.VisitDataField("insights_id", false, nil, spy.VisitUUIDDataType())}}
	err := spy.VisitResource("host", "hbi", commonMembers, reporterMembers)
	assert.NoError(t, err)

	spy.AssertJSON(t, `{
		"host": {
			"common": {"fields": [{"name": "workspace_id", "required": true, "type": {"kind": "text"}}]},
			"reporters": {
				"hbi": {"fields": [{"name": "insights_id", "required": false, "type": {"kind": "uuid"}}]}
			}
		}
	}`)
}
