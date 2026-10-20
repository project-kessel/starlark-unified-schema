load("kessel.star", "resource", "uuid")
load("extensions/rbac.star", "add_v1_based_permission")
load("cost_model/common_representation.star", common="cost_model")

cost_model_view = add_v1_based_permission("cost_management", "cost_management", "cost_model", "read", "cost_management_cost_model_view")
cost_model_edit = add_v1_based_permission("cost_management", "cost_management", "cost_model", "write", "cost_management_cost_model_edit")

cost_model = resource(
    "cost_management",
    common=common,
    id_type=uuid(),
    fields={},
    permissions={
    "view": lambda resource: cost_model_view(resource.workspace_id),
    "edit": lambda resource: cost_model_edit(resource.workspace_id),
    },
)
