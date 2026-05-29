load("kessel.star", "relation", "assignable", "cardinality", "union", "intersect", "exclude", "subref", "ref", "uuid")
load("rbac.star", "workspace")

add_member("rbac", "workspace", "child", relation(assignable("rbac", "workspace", cardinality.Any, uuid())))

add_member("rbac", "workspace", "descendents", relation(
    union(
        ref("child"),
        subref("child", "descendents")
    )
))

service = {
    "allowed_workspaces": relation(assignable("rbac", "workspace", cardinality.Any, uuid())),

    "billing_account": relation(assignable("feature", "billing_account", cardinality.Any, uuid())),

    "parent": relation(assignable("feature", "service", cardinality.Any, uuid())),

    "does_workspace_have_service_preference": relation(
        union(
            union(
                ref("allowed_workspaces"),
                subref("allowed_workspaces", "descendents")
            ),
            subref("parent", "does_workspace_have_service_preference")
        )
    ),

    "does_workspace_have_license": relation(
        subref("billing_account", "enabled_workspaces")
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

    "enabled_workspaces": relation(
        union(
            ref("workspace"),
            subref("workspace", "descendents")
        )
    )
}
