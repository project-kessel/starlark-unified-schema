relationKinds = ["AtMostOne", "ExactlyOne", "AtLeastOne", "Many", "ref", "subref"]

def _make_logic_operators(self_ref):
    """Returns dict of operator methods that can be mixed into any struct."""
    def intersect(other):
        return _make_logic_node("and", self_ref, other)
    def union(other):
        return _make_logic_node("or", self_ref, other)
    def exclude(other):
        return _make_logic_node("unless", self_ref, other)
    return intersect, union, exclude

def _make_logic_node(kind, left, right):
    """Creates and/or/unless struct with logic operators."""
    node = struct(
        kind=kind,
        left=left,
        right=right,
        intersect=None,
        union=None,
        exclude=None,
    )
    # Now create proper operators with the node as self_ref
    intersect, union, exclude = _make_logic_operators(node)
    return struct(
        kind=kind,
        left=left,
        right=right,
        intersect=intersect,
        union=union,
        exclude=exclude
    )

def _make_ref(name, field_type):
    """Creates ref struct with logic operators and subrefs."""
    attrs = {
        "kind": "ref",
        "name": name,
    }

    # Add subrefs
    for subref_name in field_type.fields:
        if field_type.kind in relationKinds:
            attrs[subref_name] = _make_subref(name, subref_name)

    # Create initial struct
    ref = struct(
        kind="ref",
        name=name,
        intersect=None, # placeholder, will be replaced
        union=None,
        exclude=None,
        **attrs
    )

    # Add logic operators
    intersect, union, exclude = _make_logic_operators(ref)
    ref.intersect = intersect
    ref.union = union
    ref.exclude = exclude

    return ref

def _make_subref(parent_name, child_name):
    """Creates subref struct with logic operators."""
    subref = struct(
        kind="subref",
        name=parent_name,
        sub=child_name,
        intersect=None, # placeholder, will be replaced
        union=None,
        exclude=None,
    )
    # Now create proper operators
    intersect, union, exclude = _make_logic_operators(subref)

    subref.intersect = intersect
    subref.union = union
    subref.exclude = exclude
    return subref

# Between this function and _create_proxy, the proxies are probably incomplete, but that's something to come back to when tests run.
def _extract_relation_and_permission_types(resource):
    """Returns dict of {relation_name: target_type(s)} for cross-type relations."""
    relation_types = {}
    for name in resource:
        rel_def = resource[name]
        if type(rel_def) == "struct" and rel_def.kind in relationKinds and "type" in dir(rel_def):
            target_type = rel_def.type

            # Skip self-references
            if type(target_type) == "dict" and target_type.get("kind") == "selfType":
                continue

            # Handle type unions
            if type(target_type) == "dict" and target_type.get("kind") == "typeUnion":
                relation_types[name] = target_type.get("types", [])
            else:
                relation_types[name] = [target_type]

    return relation_types

def _create_proxy(common, fields):
    fields_types = _extract_relation_and_permission_types(common) | _extract_relation_and_permission_types(fields)
    combined_fields = common | fields
    proxy_fields = {}

    for field_name in combined_fields:
        field = combined_fields[field_name]
        if field_name in fields_types:
            proxy_fields[field_name] = _make_ref(field_name, fields_types[field_name])

    return struct(kind="proxy", **proxy_fields)

def _process_permissions(common, object, permissions):
    proxy = _create_proxy(common, object)
    for permission_name in permissions:
        factory = permissions[permission_name]
        body = factory(proxy)
        object[permission_name] = struct(kind="permission", body=body)
    return object

def resource(reporter, id_type, common={}, fields={}, permissions={}):
    _process_permissions(common, fields, permissions)
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

def _createRelation(cardinality, type):
    return struct(
        kind="relation",
        cardinality=cardinality,
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