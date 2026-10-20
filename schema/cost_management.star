load("extensions/rbac.star", "add_v1_based_permission")

# Settings has no resource. A workspace check answers read and write.
add_v1_based_permission("cost_management", "cost_management", "settings", "read", "cost_management_settings_view")
add_v1_based_permission("cost_management", "cost_management", "settings", "write", "cost_management_settings_edit")
