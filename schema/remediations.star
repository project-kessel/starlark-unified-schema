load("extensions/rbac.star", "add_v1_based_permission")

# This is working around a conflict between V1 and V2 permission names.
# As things are right now, we can't follow the naming convention exactly because there will be conflict
# Therefore we are deviating from the ${service}_${resource}_${action} format
add_v1_based_permission("remediations", "remediation", "read", "remediations_view_remediation")
add_v1_based_permission("remediations", "remediation", "write", "remediations_edit_remediation")
add_v1_based_permission("remediations", "remediation", "execute", "remediations_execute_remediation")
