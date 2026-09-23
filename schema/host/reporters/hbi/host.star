load("kessel.star", "resource", "field", "uuid", "text", "nullable", "union")
load("extensions/rbac.star", "add_v1_based_permission")
load("host/common_representation.star", common="host")

add_v1_based_permission("hbi", "staleness", "staleness", "read", "staleness_staleness_view")
add_v1_based_permission("hbi", "staleness", "staleness", "write", "staleness_staleness_update")
add_v1_based_permission("hbi", "inventory", "views", "read", "inventory_view_view")
add_v1_based_permission("hbi", "inventory", "views", "write", "inventory_view_update")

inventory_host_view = add_v1_based_permission("hbi", "inventory", "hosts", "read", "inventory_host_view")
inventory_host_update = add_v1_based_permission("hbi", "inventory", "hosts", "write", "inventory_host_update")
inventory_host_delete = add_v1_based_permission("hbi", "inventory", "hosts", "write", "inventory_host_delete")
inventory_host_move = add_v1_based_permission("hbi", "inventory", "groups", "write", "inventory_host_move")


host = resource("hbi", common=common, 
id_type=uuid(),
fields={
    "satellite_id": field(type=nullable(union(uuid(), text(regex="^\\d{10}$")))),
    "subscription_manager_id": field(type=nullable(uuid())),
    "insights_id": field(type=nullable(uuid())),
    "ansible_host": field(type=nullable(text(maxLength=255))),
}, permissions={
    "view": lambda h: inventory_host_view(h.workspace_id),
    "update": lambda h: inventory_host_update(h.workspace_id),
    "delete": lambda h: inventory_host_delete(h.workspace_id),
    "move": lambda h: inventory_host_move(h.workspace_id)
})
