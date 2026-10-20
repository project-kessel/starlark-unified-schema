load("kessel.star", "resource", "text")
load("extensions/rbac.star", "add_v1_based_permission")
load("aws_account/common_representation.star", common="aws_account")

aws_account_view = add_v1_based_permission("cost_management", "cost_management", "aws_account", "read", "cost_management_aws_account_view")

aws_account = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={},
    permissions={
    "view": lambda resource: aws_account_view(resource.workspace_id),
    },
)
