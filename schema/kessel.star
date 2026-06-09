def self():
    return {
        "kind": "selfType"
    }

def _createRelation(kind, type):
    return {
        "kind": kind,
        "type": type,
    }

def atMostOne(type):
    return _createRelation("atMostOne", type)

def one(type):
    return _createRelation("exactlyOne", type)

def atLeastOne(type):
    return _createRelation("atLeastOne", type)

def many(type):
    return _createRelation("many", type)

def all(type):
    return _createRelation("boolean", type)

def resource_type(properties):
    for name in properties:
        prop = properties[name]
        prop["parentType"] = properties

    return properties