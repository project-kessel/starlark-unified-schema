load("extensions/rbac.star", "add_v1_based_permission", "add_contingent_permission")
load("extensions/hbi.star", "expose_host_permission")

# Lights up ros_read_analysis_assigned on workspaces where ros:analysis:read or wildcards are assigned and their descendants
add_v1_based_permission("ros", "ros", "analysis", "read", "ros_read_analysis_assigned")
# Lights up the actual ros_read_analysis permission on workspaces whose ancestry include both inventory_hosts_read and ros_read_analysis_assigned
add_contingent_permission("ros", "inventory_host_view", "ros_read_analysis_assigned", "ros_read_analysis")
# Passes through the ros_read_analysis permission from a host's workspace to the host itself
expose_host_permission("ros", "ros_read_analysis", "ros_read_analysis")
