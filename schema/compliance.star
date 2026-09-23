load("extensions/rbac.star", "add_v1_based_permission", "add_v1only_permission", "add_contingent_permission")

# Policy
add_v1_based_permission("compliance", "compliance", "policy", "read", "compliance_policy_view")
add_v1_based_permission("compliance", "compliance", "policy", "create", "compliance_policy_new")
add_v1_based_permission("compliance", "compliance", "policy", "write", "compliance_policy_edit")
add_v1_based_permission("compliance", "compliance", "policy", "delete", "compliance_policy_remove")

# These 2 permissions are not used by compliance backend - they were used by the frontend but they updated to align with the backend
# https://github.com/RedHatInsights/compliance-frontend/pull/2654/files
add_v1only_permission("compliance", "compliance_policy_update")
add_v1only_permission("compliance", "compliance_report_delete")

# Report
add_v1_based_permission("compliance", "compliance", "report", "read", "compliance_report_view")

# System
add_v1_based_permission("compliance", "compliance", "system", "read", "compliance_system_view_assigned")
add_contingent_permission("compliance", "inventory_host_view", "compliance_system_view_assigned", "compliance_system_view")
