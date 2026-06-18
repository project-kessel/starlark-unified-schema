def resource(reporter, id_type, common={}, fields={}, permissions={}):
    return struct(kind="resource", reporter=reporter, id_type=id_type, common=common, fields=fields)

def text(minLength=None, maxLength=None, regex=None):
    return struct(kind="text", minLength=minLength, maxLength=maxLength, regex=regex)

def uuid():
    return struct(kind="uuid")

def numeric_id(min=None, max=None):
    return struct(kind="numeric_id", min=min, max=max)

def boolean():
    return struct(kind="boolean")

def date_time():
    return struct(kind="date_time")

def enum(values):
    return struct(kind="enum", values=values)

def nullable(inner):
    return struct(kind="nullable", inner=inner)

def union(left, right):
    return struct(kind="union", left=left, right=right)

def array(items):
    return struct(kind="array", items=items)

def object(properties, required=[]):
    return struct(kind="object", properties=properties, required=required)

def field(type, required=False, description=None):
    return struct(kind="field", type=type, required=required, description=description)

def self():
    return {
        "kind": "selfType"
    }

def _createRelation(kind, type):
    relation = {
        "kind": kind,
        "type": type,
    }

    # Create a struct with both data fields and methods
    return struct(
        kind=kind,
        type=type,
    )

def atMostOne(type):
    return _createRelation("AtMostOne", type)

def one(type):
    return _createRelation("ExactlyOne", type)

def atLeastOne(type):
    return _createRelation("AtLeastOne", type)

def many(type):
    return _createRelation("Many", type)

def anyOf(*types):
    return {
        "kind": "typeUnion",
        "types": list(types)
    }

def any(root, *refs):
    for ref in refs:
        root = root.union(ref)
    return root

def all(root, *refs):
    for ref in refs:
        root = root.intersect(ref)
    return root