load("kessel.star", "field", "text", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

host = {
    "workspace_id": one(workspace)
}
