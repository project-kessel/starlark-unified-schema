load ("kessel.star", "resource_type", "atMostOne", "many", "self")

workspace = resource_type ({
    "parent": atMostOne(self()),
    "descendents": many(self())
})