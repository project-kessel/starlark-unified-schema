load("kessel.star", "resource", "text", "at_most_one", "self")
load("extensions/rbac.star", "add_v1_based_permission")
load("aws_organizational_unit/common_representation.star", common="aws_organizational_unit")

aws_organizational_unit_view = add_v1_based_permission("cost_management", "cost_management", "aws_organizational_unit", "read", "cost_management_aws_organizational_unit_view")

aws_organizational_unit = resource(
    "cost_management",
    common=common,
    id_type=text(),
    fields={
        # Membership only. The top organizational unit has no parent. view follows workspace_id.
        "parent_ou": at_most_one(self()),
    },
    permissions={
        "view": lambda resource: aws_organizational_unit_view(resource.workspace_id),
    },
)
