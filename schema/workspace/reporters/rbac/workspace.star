load("kessel.star", "resource", "at_most_one", "many", "self", "uuid")

reporter("rbac")

workspace = resource(id_type=uuid(),
fields={
    "parent": at_most_one(self()),
})
