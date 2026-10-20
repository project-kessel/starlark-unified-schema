load("kessel.star", "resource", "text")
load("extensions/rbac.star", "add_v1_based_permission")
load("azure_subscription/common_representation.star", common="azure_subscription")

azure_subscription_view = add_v1_based_permission("cost_management", "cost_management", "azure_subscription_guid", "read", "cost_management_azure_subscription_view")

azure_subscription = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={},
    permissions={
    "view": lambda resource: azure_subscription_view(resource.workspace_id),
    },
)
