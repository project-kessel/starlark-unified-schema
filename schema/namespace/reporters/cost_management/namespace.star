load("kessel.star", "resource", "text", "one")
load("extensions/rbac.star", "add_v1_based_permission")
load("namespace/common_representation.star", common="namespace")
load("ocp_cluster/reporters/cost_management/ocp_cluster.star", "ocp_cluster")

# Openshift project. A namespace belongs to one cluster and can run on many nodes.
namespace_view = add_v1_based_permission("cost_management", "cost_management", "openshift_project", "read", "cost_management_namespace_view")

namespace = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={
        # Membership only. view follows workspace_id, not cluster.
        "cluster": one(ocp_cluster),
    },
    permissions={
    "view": lambda resource: namespace_view(resource.workspace_id),
    },
)
