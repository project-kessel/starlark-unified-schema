load ("kessel.star", "resource", "atMostOne", "many", "self", "uuid")

workspace = resource (reporter="rbac", id_type=uuid(), 
fields={
    "parent": atMostOne(self()),
    "children": many(self())
},
permissions={
    "descendants": lambda w: w.children.union(w.children.descendants)
})
