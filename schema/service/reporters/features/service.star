load("kessel.star", "resource", "field", "uuid", "many", "atMostOne", "self", "any")
load("workspace/reporters/rbac/workspace.star", "workspace")
load("billing_account/reporters/features/billing_account.star", "billing_account")

service = resource(reporter="features",
id_type=uuid(),
fields={
    "allowed_workspaces": many(workspace),
    "billing_account": many(billing_account),
    "parent": atMostOne(self())
},
permissions={ 
    "does_workspace_have_service_preference": lambda s: any(
        s.allowed_workspaces,
        s.allowed_workspaces.descendants,
        s.parent.does_workspace_have_service_preference
    ),

    "does_workspace_have_license": lambda s: s.billing_account.enabled_workspaces,

    "can_workspace_use_service": lambda s: s.does_workspace_have_service_preference.intersect(s.does_workspace_have_license)
})