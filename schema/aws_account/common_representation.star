load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

aws_account = {
    "workspace_id": one(workspace),
}
