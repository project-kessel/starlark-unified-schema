load("kessel.star", "resource", "text", "one")
load("extensions/rbac.star", "add_v1_based_permission")
load("ocp_node/common_representation.star", common="ocp_node")
load("ocp_cluster/reporters/cost_management/ocp_cluster.star", "ocp_cluster")

ocp_node_view = add_v1_based_permission("cost_management", "cost_management", "openshift_node", "read", "cost_management_ocp_node_view")

ocp_node = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={
        # Membership only. view follows workspace_id, not cluster.
        "cluster": one(ocp_cluster),
    },
    permissions={
    "view": lambda resource: ocp_node_view(resource.workspace_id),
    },
)
