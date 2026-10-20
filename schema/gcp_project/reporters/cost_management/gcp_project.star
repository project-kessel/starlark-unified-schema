load("kessel.star", "resource", "text")
load("extensions/rbac.star", "add_v1_based_permission")
load("gcp_project/common_representation.star", common="gcp_project")

gcp_project_view = add_v1_based_permission("cost_management", "cost_management", "gcp_project", "read", "cost_management_gcp_project_view")

gcp_project = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={},
    permissions={
    "view": lambda resource: gcp_project_view(resource.workspace_id),
    },
)
