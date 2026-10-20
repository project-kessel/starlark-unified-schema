load("kessel.star", "resource", "text")
load("extensions/rbac.star", "add_v1_based_permission")
load("gcp_account/common_representation.star", common="gcp_account")

gcp_account_view = add_v1_based_permission("cost_management", "cost_management", "gcp_account", "read", "cost_management_gcp_account_view")

gcp_account = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={},
    permissions={
    "view": lambda resource: gcp_account_view(resource.workspace_id),
    },
)
