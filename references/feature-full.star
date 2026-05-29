load("kessel.star", "relation", "assignable", "cardinality", "union", "intersect", "exclude", "subref", "ref", "uuid")
load("rbac.star", "workspace")

add_member("rbac", "workspace", "inherited_service_preference_from_parent", relation(assignable("rbac", "workspace", cardinality.Any, uuid())))

add_member("rbac", "workspace", "inherited_billing_account_from_parent", relation(assignable("rbac", "workspace", cardinality.Any, uuid())))

add_member("rbac", "workspace", "child", relation(assignable("rbac", "workspace", cardinality.Any, uuid())))

add_member("rbac", "workspace", "inheriting_preference_children", relation(
    subref("inherited_service_preference_from_parent", "all_preference_inheriting_children_workspaces")
))

add_member("rbac", "workspace", "all_preference_inheriting_children_workspaces", relation(
    union(
        subref("child", "inherited_service_preference_from_parent"),
        subref("child", "inheriting_preference_children")
    )
))

add_member("rbac", "workspace", "inheriting_billing_children", relation(
    subref("inherited_billing_account_from_parent", "all_billing_inheriting_children_workspaces")
))

add_member("rbac", "workspace", "all_billing_inheriting_children_workspaces", relation(
    union(
        subref("child", "inherited_billing_account_from_parent"),
        subref("child", "inheriting_billing_children")
    )
))

service = {
    "allowed_workspaces": relation(assignable("rbac", "workspace", cardinality.Any, uuid())),

    "denied_workspaces": relation(assignable("rbac", "workspace", cardinality.Any, uuid())),

    "billing_account": relation(assignable("feature", "billing_account", cardinality.Any, uuid())),

    "parent": relation(assignable("feature", "service", cardinality.Any, uuid())),

    "does_workspace_have_service_preference": relation(
        exclude(
            union(
                union(
                    exclude(
                        subref("parent", "does_workspace_have_service_preference"),
                        subref("denied_workspaces", "all_preference_inheriting_children_workspaces")
                    ),
                    ref("allowed_workspaces")
                ),
                subref("allowed_workspaces", "all_preference_inheriting_children_workspaces")
            ),
            ref("denied_workspaces")
        )
    ),

    "does_workspace_have_license": relation(
        subref("billing_account", "workspaces_that_are_using_this_billing_account")
    ),

    "can_workspace_use_service": relation(
        intersect(
            ref("does_workspace_have_service_preference"),
            ref("does_workspace_have_license")
        )
    )
}

billing_account = {
    "workspace": relation(assignable("rbac", "workspace", cardinality.Any, uuid())),

    "workspaces_that_are_using_this_billing_account": relation(
        union(
            ref("workspace"),
            subref("workspace", "all_billing_inheriting_children_workspaces")
        )
    )
}
