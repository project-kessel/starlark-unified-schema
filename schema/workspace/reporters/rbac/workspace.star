load ("kessel.star", "resource", "at_most_one", "many", "self", "uuid")

workspace = resource (reporter="rbac", id_type=uuid(), 
fields={
    "parent": at_most_one(self()),
})
