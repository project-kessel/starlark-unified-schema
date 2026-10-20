load("kessel.star", "resource", "text", "at_most_one")
load("extensions/rbac.star", "add_v1_based_permission")
load("aws_account/common_representation.star", common="aws_account")
load("aws_organizational_unit/reporters/cost_management/aws_organizational_unit.star", "aws_organizational_unit")

aws_account_view = add_v1_based_permission("cost_management", "cost_management", "aws_account", "read", "cost_management_aws_account_view")

aws_account = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={
        # Membership only. An account with no organizational unit omits this. view follows workspace_id.
        "organizational_unit": at_most_one(aws_organizational_unit),
    },
    permissions={
        "view": lambda resource: aws_account_view(resource.workspace_id),
    },
)
