"""The Kessel schema DSL standard library."""

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
    sub = lambda n: _make_subref(self_ref_holder[0].name, n)
    ref = struct(kind="ref", name=name, sub=sub, intersect=intersect, union=union, exclude=exclude, **subrefs)
    
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
            #TODO: will need to traverse the parent type as well once references to sub-types are supported
        else:
            fail("Unknown relation type: {}".format(type.kind))
    return result.keys()

def _create_proxy(common, parent, fields, permission_names):
    fields_types = _extract_relation_types(common) | _extract_relation_types(fields)

    if parent != None:
        fields_types = fields_types | _extract_relation_types(parent.fields) | _extract_relation_types(parent.common)
        for name in parent.fields:
            p = parent.fields[name]
            if type(p) == "struct" and p.kind == "permission":
                permission_names.append(name) #Include permission names from the parent type

    self_names = fields_types.keys() + permission_names
    proxy_fields = {}

    for field_name in fields_types:
        proxy_fields[field_name] = _make_ref(field_name, _get_relation_names(self_names, fields_types[field_name]))

    for permission_name in permission_names:
        proxy_fields[permission_name] = _make_ref(permission_name, [])


    return struct(kind="proxy", **proxy_fields)

def _process_permissions(common, parent, object, permissions):
    proxy = _create_proxy(common, parent, object, permissions.keys())
    combined_fields = {}

    for field_name in object:
        combined_fields[field_name] = object[field_name]

    for permission_name in permissions:
        factory = permissions[permission_name]
        body = factory(proxy)
        combined_fields[permission_name] = struct(kind="permission", body=body)

    return combined_fields

def resource(reporter, id_type=None, extends=None, final=False, common={}, fields={}, permissions={}):
    """Declares a resource type.

    Each resource reported to Kessel must be of a type. A type has fields and permissions.
    Fields are the data elements a service reports about a resource. Permissions are like
    computed fields that determine which subjects would be allowed to perform an operation.
    Permissions can reference a subset of data fields (references to other resources, booleans)
    and use set operations to express more complex authorization logic (see Permission expressions below.)

    Args:
      reporter: name of the reporter that reports this resource, e.g. `"hbi"`.
      id_type: data type of the resource's identifier. Must be a value data type (not another resource type.)
        Required unless `extends` is given, in which case the parent's id type is
        inherited and passing one here is an error.
      extends: another resource whose common members, fields, and permissions this
        resource inherits. An extending resource has no independent identity and must always be
        associated with an instance of the parent type (reported with the same id.)  Only one level of inheritance is supported, so an
        extending resource is implicitly final.
      final: if true, prevents other resource types from extending this one. This is recommended for types that aren't likely
        to be part of another service's domain or have no stable identity. Forced to True when `extends` is given.
      common: reusable dict of members to add to this type. Types that represent different domain-specific views
        of the same real-world object (ex: a RHEL VM) can reference the same block of members to have common fields.
        This is meant for future use with correlation.
      fields: dict of member name to declaration. Each member is either a value data field built with `field()` 
        or a relation built with a cardinality function such as `one()`, `many()`, or `boolean()`.
      permissions: dict of permission name to a lambda that takes the resource and returns a permission expression, 
        described below. A name starting with an underscore declares an internal member that other permissions
        in the same type can build on.

    Permission expressions:

    Each permission lambda receives an object representing the resource being
    defined. This object has properties representing the relationship properties of the
    type being defined, including common and inherited ones. Relationship properties in turn
    have the applicable properties from the types they reference. Sub-properties have no properties-
    if deeper nesting is required, there needs to be another permission that continues the chain.
    Permissions also have no sub-properties - they must already represent the relevant subject type.

    References provide three chaining methods, each returning a new expression
    that can be chained again:

    - `.union(other)` — subjects from either side have access
    - `.intersect(other)` — a subject must appear on both sides to have access
    - `.exclude(other)` — a subject has if on the left and not on the right

    `any()` and `all()` are helpers for chains of `.union` and `.intersect`
    respectively.

    ```python
    workspace = resource("features", extends=rbac_workspace, fields={
        "direct_billing_account": at_most_one(billing_account),
        "ignore_inherited_paid_services": boolean(service),
    }, permissions={
        "_paid_services": lambda w: any(
            w.direct_billing_account.services,
            w.parent._paid_services.exclude(w.ignore_inherited_paid_services),
        ),
        "enabled_services": lambda w: w._paid_services.intersect(w._desired_services),
    })

    Returns:

    A resource type object
    ```
    """
    if id_type == None and extends==None:
        fail("resource: either an id_type or a type to extend must be specified")

    if extends != None:
        final=True #Only allow one level of inheritance
        if id_type != None:
            fail("resource: a resource type extending another type must not specify an id_type")
        id_type = extends.id_type #Inherit the id type from the parent

    combined_fields = _process_permissions(common, extends, fields, permissions)
    return struct(kind="resource", reporter=reporter, id_type=id_type, parent=extends, final=final, common=common, fields=combined_fields)

def text(minLength=None, maxLength=None, regex=None):
    """A text value data type, optionally length- and pattern-constrained.

    Args:
      minLength: minimum number of characters. There is no lower bound if not specified.
      maxLength: maximum number of characters. There is no upper bound if not specified.
      regex: pattern the value must match, as a regular expression. It does not apply if not specified.
        Remember that Starlark strings have no raw-string form, so backslashes
        must be doubled: `text(regex="^\\\\d{10}$")`.
    """
    return struct(kind="text", minLength=minLength, maxLength=maxLength, regex=regex)

def uuid():
    """A UUID value data type."""
    return struct(kind="uuid")

def numeric_id(min=None, max=None):
    """An integer identifier value data type, optionally range-constrained.

    Args:
      min: smallest permitted value, inclusive, or None for no lower bound.
      max: largest permitted value, inclusive, or None for no upper bound.
    """
    return struct(kind="numeric_id", min=min, max=max)

def boolean(target=None):
    """A boolean value data type, or a boolean-like relation type.

    With no argument this is a true/false data field and not usable in permissions. Given a subject
    resource type, it becomes a boolean-like relation, still reported as a true/false value but also
    usable in permissions for that subject type.

    Args:
      target: the subject type permissions the value will be used in apply to, or None for a plain
        boolean data type.
    """
    if target == None:
        return struct(kind="boolean")
    return struct(kind="relation", cardinality="All", type=target, input=struct(kind="boolean"))

def date_time():
    """A timestamp value data type. Reported as a string in a recognized datetime format."""
    return struct(kind="date_time")

def enum(values):
    """A string data type restricted to a fixed set of values.

    Args:
      values: list of permitted strings, in the order they should be published.
    """
    return struct(kind="enum", values=values)

def nullable(inner):
    """Permits null in addition to whatever `inner` allows.

    Args:
      inner: the data type to widen. Nesting a `union()` is flattened, so
        `nullable(union(uuid(), text()))` offers three alternatives rather than a
        null alongside a nested choice.
    """
    return struct(kind="nullable", inner=inner)

def union(left, right):
    """A choice between two data types; a value matching either is valid.

    Nest calls for more than two alternatives. Note this is specifically for data types. 
    See `resource` for details on boolean logic in permissions.

    Args:
      left: the first permitted data type.
      right: the second permitted data type.
    """
    return struct(kind="union", left=left, right=right)

def array(items):
    """A list data type whose elements all share one type.

    Args:
      items: the value data type of each element.
    """
    return struct(kind="array", items=items)

def object(properties, required=[]):
    """A nested structured data type.

    Args:
      properties: dict of property name to data type.
      required: list of property names that must be present. Names omitted here
        are optional.
    """
    return struct(kind="object", properties=properties, required=required)

def field(type, required=False, description=None):
    """Declares a data field on a resource, wrapping a value data type.

    Relations are declared with a cardinality helper instead and are not wrapped
    in `field()`.

    Args:
      type: the field's data type, e.g. `uuid()` or `nullable(text())`.
      required: whether a representation must carry this field.
      description: documenation for reporters reporting the field.
    """
    return struct(kind="field", type=type, required=required, description=description)

def self():
    """Refers to the resource type currently being defined.

    Use it as a relation target so a resource can point at its own type, as in
    `at_most_one(self())` for a parent link. Needed because the resource global
    does not exist yet while its own `resource()` call is being evaluated.
    """
    return struct(kind="self")

def _create_relation(cardinality, type):
    return struct(
        kind="relation",
        cardinality=cardinality,
        type=type,
    )

def wildcard(type):
    """A relation to every instance of a type at once.

    The reporter supplies a wildcard marker rather than identifiers. Use
    `boolean(type)` instead to report the same relation as a true/false flag.

    Args:
      type: the target resource or `self()`.
    """
    return _create_relation("All", type)

def at_most_one(type):
    """An optional relation to a single instance, reported as one identifier.

    Args:
      type: the target resource or `self()`
    """
    return _create_relation("AtMostOne", type)

def one(type):
    """A required relation to exactly one instance, reported as one identifier.

    Args:
      type: the target resource or `self()`.
    """
    return _create_relation("ExactlyOne", type)

def at_least_one(type):
    """A required relation to one or more instances, reported as a non-empty list.

    Args:
      type: the target resource or `self()`.
    """
    return _create_relation("AtLeastOne", type)

def many(type):
    """An optional relation to any number of instances, reported as a list.

    Args:
      type: the target resource or `self()`
    """
    return _create_relation("Many", type)

def any_of(*types):
    """***Reserved for Future Use***"""
    return struct(kind="typeUnion", types=list(types))

def any(root, *refs):
    """Variadic function that evaluates to true when any of the given permission expressions evaluate to true.

    any(a, b, c) behaves the same as a.union(b).union(c)

    Args:
      root: the first expression.
      *refs: further expressions.
    """
    for ref in refs:
        root = root.union(ref)
    return root

def all(root, *refs):
    """Variadic function that evaluates to true only when all of the given permission evaluate to true.

    all(a, b, c) behaves the same as a.intersect(b).intersect(c)

    Args:
      root: the first expression.
      *refs: further expressions.
    """
    for ref in refs:
        root = root.intersect(ref)
    return root
