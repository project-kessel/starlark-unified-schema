load("kessel.star", "many", "atMostOne", "any", "permissions", "self")
load("rbac.star", "workspace")

billing_account = {
    "workspace": many(workspace)
}

permissions(billing_account, {
    "enabled_workspaces": lambda b: b.workspace.union(b.workspace.descendants)
})

service = {
    "allowed_workspaces": many(workspace),
    "billing_account": many(billing_account),
    "parent": atMostOne(self())
}

permissions(service,
{ 
    "does_workspace_have_service_preference": lambda s: any(
        s.allowed_workspaces,
        s.allowed_workspaces.descendants,
        s.parent.does_workspace_have_service_preference
    ),

    "does_workspace_have_license": lambda s: s.billing_account.enabled_workspaces,

    "can_workspace_use_service": lambda s: s.does_workspace_have_service_preference.intersect(s.does_workspace_have_license)
})