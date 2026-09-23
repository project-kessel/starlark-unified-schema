load("extensions/rbac.star", "add_v1_based_permission")

add_v1_based_permission("config_manager", "config_manager", "profile", "read", "config_manager_profile_view")
add_v1_based_permission("config_manager", "config_manager", "profile", "write", "config_manager_profile_edit")

add_v1_based_permission("config_manager", "config_manager", "activation_keys", "read", "config_manager_activation_keys_view")
add_v1_based_permission("config_manager", "config_manager", "activation_keys", "write", "config_manager_activation_keys_edit")
