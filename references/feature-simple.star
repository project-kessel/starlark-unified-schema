# Feature Simple Schema — Original proposal (service → workspace direction)
#
# Service points to workspaces and billing accounts. Workspace has children/descendants
# for downward traversal. Check starts at the service:
#     zed permission check features/service:RHEL can_workspace_use_service rbac/workspace:ws1
#
# This is the same schema as feature-simple-old.star, translated to the new syntax
# used by schema/kessel.star. See feature-simple-old.star for the original syntax.

load("kessel.star", "resource", "uuid", "many", "at_most_one", "self", "any")

workspace = resource(reporter="rbac", id_type=uuid(),
fields={
    "parent": at_most_one(self()),
    "children": many(self()),
},
permissions={
    "descendants": lambda w: w.children.union(w.children.descendants),
})

billing_account = resource(reporter="features", id_type=uuid(),
fields={
    "workspaces": many(workspace),
},
permissions={
    "enabled_workspaces": lambda b: b.workspaces.union(b.workspaces.descendants),
})

service = resource(reporter="features", id_type=uuid(),
fields={
    "allowed_workspaces": many(workspace),
    "billing_account": many(billing_account),
    "parent": at_most_one(self()),
},
permissions={
    "does_workspace_have_service_preference": lambda s: any(
        s.allowed_workspaces,
        s.allowed_workspaces.descendants,
        s.parent.does_workspace_have_service_preference,
    ),

    "does_workspace_have_license": lambda s: s.billing_account.enabled_workspaces,

    "can_workspace_use_service": lambda s: s.does_workspace_have_service_preference.intersect(
        s.does_workspace_have_license,
    ),
})
