load("kessel.star", "resource", "field", "uuid", "text", "nullable", "union")
load("host/common_representation.star", common="host")

host = resource("hbi", common=common, 
id_type=uuid(),
fields={
    "satellite_id": field(type=nullable(union(uuid(), text(regex="^\\d{10}$")))),
    "subscription_manager_id": field(type=nullable(uuid())),
    "insights_id": field(type=nullable(uuid())),
    "ansible_host": field(type=nullable(text(maxLength=255))),
})
