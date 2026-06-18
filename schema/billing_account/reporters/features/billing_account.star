load("kessel.star", "resource", "field", "uuid", "permissions", "many")

billing_account = resource(reporter="features", 
id_type=uuid(), 
fields={
    "workspaces": many(),
}, permissions={
    "enabled_workspaces": lambda b: b.workspaces.union(b.workspaces.descendants)
})
