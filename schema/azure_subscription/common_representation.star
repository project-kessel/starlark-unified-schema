load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

azure_subscription = {
    "workspace_id": one(workspace),
}
