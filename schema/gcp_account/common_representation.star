load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

gcp_account = {
    "workspace_id": one(workspace),
}
