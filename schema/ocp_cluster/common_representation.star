load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

ocp_cluster = {
    "workspace_id": one(workspace),
}
