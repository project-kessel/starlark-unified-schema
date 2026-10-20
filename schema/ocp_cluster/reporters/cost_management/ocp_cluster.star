load("kessel.star", "resource", "text")
load("extensions/rbac.star", "add_v1_based_permission")
load("ocp_cluster/common_representation.star", common="ocp_cluster")

ocp_cluster_view = add_v1_based_permission("cost_management", "cost_management", "openshift_cluster", "read", "cost_management_ocp_cluster_view")

ocp_cluster = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={},
    permissions={
    "view": lambda resource: ocp_cluster_view(resource.workspace_id),
    },
)
