load("rbac", "workspace")

billing_account = {
    "workspace": many(workspace)
}

permissions(billing_account, {
    "enabled_workspaces": lambda(b): b.workspace.or(b.workspace.descendents)
})

service = {
    "allowed_workspaces": many(workspace)
    "billing_account": many(billing_account)
    "parent": atMostOne(self())
}

permissions(service,
{ 
    "does_workspace_have_service_preference": lambda(s): any(
        s.allowed_workspaces,
        s.allowed_workspaces.descendents,
        s.parent.does_workspace_have_service_preference
    ),

    "does_workspace_have_license": lambda(s): s.billing_account.enabled_workspaces,

    "can_workspace_use_service": lambda(s): s.does_workspace_have_service_preference.and(s.does_workspace_have_license)
})