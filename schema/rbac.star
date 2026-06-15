load ("kessel.star", "permissions", "resource_type", "atMostOne", "many", "self")

workspace = resource_type ({
    "parent": atMostOne(self()),
    "children": many(self())
})

permissions(workspace, {
    "descendants": lambda w: w.children.union(w.children.descendants)
})