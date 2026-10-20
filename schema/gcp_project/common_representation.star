load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

gcp_project = {
    "workspace_id": one(workspace),
}
