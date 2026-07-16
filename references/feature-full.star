# Feature Full Schema — Original proposal with deny lists and selective inheritance
#
# Same direction as feature-simple (service → workspace), but adds:
#   - Deny lists (denied_workspaces) with exclude operations
#   - Selective inheritance: workspaces can opt in/out of inheriting
#     service preferences and billing accounts separately
#   - Workspace has children + selective inheritance tracking
#
# Check: zed permission check features/service:RHEL can_workspace_use_service rbac/workspace:ws1
#
# This is the same schema as feature-full-old.star, translated to the new syntax.
# See feature-full-old.star for the original syntax.

load("kessel.star", "resource", "uuid", "many", "at_most_one", "self", "any")

workspace = resource(reporter="rbac", id_type=uuid(),
fields={
    "parent": at_most_one(self()),
    "children": many(self()),
    "inherited_service_preference_from_parent": many(self()),
    "inherited_billing_account_from_parent": many(self()),
},
permissions={
    "inheriting_preference_children": lambda w: w.inherited_service_preference_from_parent.all_preference_inheriting_children_workspaces,

    "all_preference_inheriting_children_workspaces": lambda w: w.children.inherited_service_preference_from_parent.union(
        w.children.inheriting_preference_children,
    ),

    "inheriting_billing_children": lambda w: w.inherited_billing_account_from_parent.all_billing_inheriting_children_workspaces,

    "all_billing_inheriting_children_workspaces": lambda w: w.children.inherited_billing_account_from_parent.union(
        w.children.inheriting_billing_children,
    ),
})

billing_account = resource(reporter="features", id_type=uuid(),
fields={
    "workspaces": many(workspace),
},
permissions={
    "workspaces_that_are_using_this_billing_account": lambda b: b.workspaces.union(
        b.workspaces.all_billing_inheriting_children_workspaces,
    ),
})

service = resource(reporter="features", id_type=uuid(),
fields={
    "allowed_workspaces": many(workspace),
    "denied_workspaces": many(workspace),
    "billing_account": many(billing_account),
    "parent": at_most_one(self()),
},
permissions={
    "does_workspace_have_service_preference": lambda s:
        s.parent.does_workspace_have_service_preference
            .exclude(s.denied_workspaces.all_preference_inheriting_children_workspaces)
            .union(s.allowed_workspaces)
            .union(s.allowed_workspaces.all_preference_inheriting_children_workspaces)
            .exclude(s.denied_workspaces),

    "does_workspace_have_license": lambda s: s.billing_account.workspaces_that_are_using_this_billing_account,

    "can_workspace_use_service": lambda s: s.does_workspace_have_service_preference.intersect(
        s.does_workspace_have_license,
    ),
})
