load("extensions/rbac.star", "add_v1_based_permission")

add_v1_based_permission("tasks", "task", "read", "tasks_task_view")
add_v1_based_permission("tasks", "task", "write", "tasks_task_edit")
