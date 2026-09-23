load("extensions/rbac.star", "add_v1_based_permission", "add_v1only_permission", "add_contingent_permission")
load("extensions/hbi.star", "expose_host_permission")

add_v1_based_permission("patch", "patch", "system", "read", "patch_system_view_assigned")
add_contingent_permission("patch", "inventory_host_view", "patch_system_view_assigned", "patch_system_view")
expose_host_permission("patch", "patch_system_view", "patch_system_view")

add_v1_based_permission("patch", "patch", "system", "write", "patch_system_edit_assigned")
add_contingent_permission("patch", "inventory_host_view", "patch_system_edit_assigned", "patch_system_edit")
expose_host_permission("patch", "patch_system_edit", "patch_system_edit")

add_v1only_permission("patch", "patch_template_write")

add_contingent_permission("patch", "patch_system_view", "content_sources_template_view", "patch_template_view")
expose_host_permission("patch", "patch_template_view", "patch_template_view")

add_contingent_permission("patch", "patch_system_edit", "content_sources_template_edit", "patch_template_edit")
expose_host_permission("patch", "patch_template_edit", "patch_template_edit")
