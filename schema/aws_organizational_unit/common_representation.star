load("kessel.star", "one")
load("workspace/reporters/rbac/workspace.star", "workspace")

aws_organizational_unit = {
    "workspace_id": one(workspace),
}
