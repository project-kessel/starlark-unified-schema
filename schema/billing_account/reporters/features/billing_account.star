load("kessel.star", "resource", "field", "uuid", "many")
load("workspace/reporters/rbac/workspace.star", "workspace")

billing_account = resource(reporter="features", 
id_type=uuid(), 
fields={
    "workspaces": many(workspace),
}, permissions={
    "enabled_workspaces": lambda b: b.workspaces.union(b.workspaces.descendants)
})
