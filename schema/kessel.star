relationKinds = ["AtMostOne", "ExactlyOne", "AtLeastOne", "Many", "ref", "subref"]

def _make_logic_operators(self_ref_holder):
    """Constructs operator methods that can be mixed into any struct."""
    # self_ref_holder is a list with a single element - this is meant to mimic struct pointer pointer
    # That way the struct pointer can be modified after this method is called
    def intersect(other):
        return _make_logic_node("and", self_ref_holder[0], other)
    def union(other):
        return _make_logic_node("or", self_ref_holder[0], other)
    def exclude(other):
        return _make_logic_node("unless", self_ref_holder[0], other)

    return intersect, union, exclude

def _make_logic_node(kind, left, right):
    """Creates and/or/unless struct with logic operators."""
    self_ref_holder = [None]

    intersect, union, exclude = _make_logic_operators(self_ref_holder)

    node = struct(
        kind=kind,
        left=left,
        right=right,
        intersect=intersect,
        union=union,
        exclude=exclude,
    )
    
    self_ref_holder[0] = node
    return node

def _make_ref(name, child_names):
    """Creates ref struct with logic operators and subrefs."""
    self_ref_holder = [None]
    # Add subrefs
    subrefs = {}
    for subref_name in child_names:
        subrefs[subref_name] = _make_subref(name, subref_name)

    intersect, union, exclude = _make_logic_operators(self_ref_holder)
    ref = struct(kind="ref", name=name, intersect=intersect, union=union, exclude=exclude, **subrefs)
    
    self_ref_holder[0] = ref
    return ref

def _make_subref(parent_name, child_name):
    """Creates subref struct with logic operators."""
    self_ref_holder = [None]
    intersect, union, exclude = _make_logic_operators(self_ref_holder)
    subref = struct(
        kind="subref",
        name=parent_name,
        sub=child_name,
        intersect=intersect,
        union=union,
        exclude=exclude,
    )
    # Now create proper operators
    self_ref_holder[0] = subref
    return subref

# Between this function and _create_proxy, the proxies are probably incomplete, but that's something to come back to when tests run.
def _extract_relation_types(resource):
    """Returns dict of {relation_name: target_type(s)} for cross-type relations."""
    relation_types = {}
    for name in resource:
        rel_def = resource[name]
        if type(rel_def) == "struct" and rel_def.kind == "relation" and "type" in dir(rel_def):
            target_type = rel_def.type

            if target_type.kind == "self" or target_type.kind == "resource":
                relation_types[name] = [target_type]
            elif target_type.kind == "typeUnion":
                relation_types[name] = target_type.types
            else:
                fail("Unknown relation type: {}".format(target_type.kind))

    return relation_types

def _get_relation_names(self_names, types):
    result = {}
    for type in types:
        if type.kind == "self":
            for name in self_names:
                result[name] = True
        elif type.kind == "resource":
            for name in type.fields:
                result[name] = True
            for name in type.common:
                result[name] = True
        else:
            fail("Unknown relation type: {}".format(type.kind))
    return result.keys()

def _create_proxy(common, fields, permission_names):
    fields_types = _extract_relation_types(common) | _extract_relation_types(fields)
    self_names = fields_types.keys() + permission_names
    proxy_fields = {}

    for field_name in fields_types:
        proxy_fields[field_name] = _make_ref(field_name, _get_relation_names(self_names, fields_types[field_name]))

    for permission_name in permission_names:
        proxy_fields[permission_name] = _make_ref(permission_name, [])

    return struct(kind="proxy", **proxy_fields)

def _process_permissions(common, object, permissions):
    proxy = _create_proxy(common, object, permissions.keys())
    combined_fields = {}

    for field_name in object:
        combined_fields[field_name] = object[field_name]

    for permission_name in permissions:
        factory = permissions[permission_name]
        body = factory(proxy)
        combined_fields[permission_name] = struct(kind="permission", body=body)

    return combined_fields

def resource(reporter, id_type, common={}, fields={}, permissions={}):
    combined_fields = _process_permissions(common, fields, permissions)
    return struct(kind="resource", reporter=reporter, id_type=id_type, common=common, fields=combined_fields)

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
    return struct(kind="self")

def _create_relation(cardinality, type):
    return struct(
        kind="relation",
        cardinality=cardinality,
        type=type,
    )

def wildcard(type):
    return _create_relation("All", type)

def at_most_one(type):
    return _create_relation("AtMostOne", type)

def one(type):
    return _create_relation("ExactlyOne", type)

def at_least_one(type):
    return _create_relation("AtLeastOne", type)

def many(type):
    return _create_relation("Many", type)

def any_of(*types):
    return struct(kind="typeUnion", types=list(types))

def any(root, *refs):
    for ref in refs:
        root = root.union(ref)
    return root

def all(root, *refs):
    for ref in refs:
        root = root.intersect(ref)
    return root