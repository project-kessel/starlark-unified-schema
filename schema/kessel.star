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
    return _createRelation("atMostOne", type)

def one(type):
    return _createRelation("exactlyOne", type)

def atLeastOne(type):
    return _createRelation("atLeastOne", type)

def many(type):
    return _createRelation("many", type)

def boolean(type):
    return _createRelation("boolean", type)

def resource_type(properties):
    return properties

def anyOf(*types):
    return {
        "kind": "typeUnion",
        "types": list(types)
    }

# Helper functions for permissions()

def _get_property_kind(prop):
    """Returns 'relation' if prop is a relation, 'other' otherwise."""
    if type(prop) == "struct" and "kind" in dir(prop):
        if prop.kind in ["atMostOne", "exactlyOne", "atLeastOne", "many", "boolean"]:
            return "relation"
    return "other"

def _extract_relation_types(resource):
    """Returns dict of {relation_name: target_type(s)} for cross-type relations."""
    relation_types = {}
    for rel_name in resource:
        rel_def = resource[rel_name]
        if type(rel_def) == "struct" and "type" in dir(rel_def):
            target_type = rel_def.type

            # Skip self-references
            if type(target_type) == "dict" and target_type.get("kind") == "selfType":
                continue

            # Handle type unions
            if type(target_type) == "dict" and target_type.get("kind") == "typeUnion":
                relation_types[rel_name] = target_type.get("types", [])
            else:
                relation_types[rel_name] = [target_type]

    return relation_types

def _make_logic_operators(self_ref):
    """Returns dict of operator methods that can be mixed into any struct."""
    def intersect(other):
        return _make_logic_node("and", self_ref, other)
    def union(other):
        return _make_logic_node("or", self_ref, other)
    def exclude(other):
        return _make_logic_node("unless", self_ref, other)
    return {
        "intersect": intersect,
        "union": union,
        "exclude": exclude
    }

def _make_logic_node(kind, left, right):
    """Creates and/or/unless struct with logic operators."""
    node = struct(
        kind=kind,
        left=left,
        right=right,
        **_make_logic_operators(None)  # Placeholder, will be replaced
    )
    # Now create proper operators with the node as self_ref
    operators = _make_logic_operators(node)
    return struct(
        kind=kind,
        left=left,
        right=right,
        intersect=operators["intersect"],
        union=operators["union"],
        exclude=operators["exclude"]
    )

def _make_ref(name, subrefs_dict):
    """Creates ref struct with logic operators and subrefs."""
    attrs = {
        "kind": "ref",
        "name": name,
    }

    # Add subrefs
    for subref_name in subrefs_dict:
        attrs[subref_name] = _make_subref(name, subref_name)

    # Create initial struct
    ref = struct(**attrs)

    # Add logic operators
    operators = _make_logic_operators(ref)
    attrs["intersect"] = operators["intersect"]
    attrs["union"] = operators["union"]
    attrs["exclude"] = operators["exclude"]

    return struct(**attrs)

def _make_subref(parent_name, child_name):
    """Creates subref struct with logic operators."""
    subref = struct(
        kind="subref",
        name=parent_name,
        sub=child_name,
        **_make_logic_operators(None)  # Placeholder
    )
    # Now create proper operators
    operators = _make_logic_operators(subref)
    return struct(
        kind="subref",
        name=parent_name,
        sub=child_name,
        intersect=operators["intersect"],
        union=operators["union"],
        exclude=operators["exclude"]
    )

def _build_accessor_struct_with_permissions(props, relation_types, permission_names):
    """Builds accessor struct from properties and their types, including forward refs to permissions."""
    attrs = {}

    # Collect all property names for same-type subrefs (relations + future permissions)
    all_prop_names = list(props.keys())
    for perm_name in permission_names:
        all_prop_names.append(perm_name)

    for prop_name in props:
        # Collect subrefs for this property
        subrefs = {}

        # Add cross-type subrefs (include ALL properties from target type, not just relations)
        if prop_name in relation_types:
            for target_type in relation_types[prop_name]:
                if type(target_type) == "dict":
                    for target_prop_name in target_type:
                        subrefs[target_prop_name] = True

        # Add same-type subrefs (both existing relations and future permissions)
        for other_prop_name in all_prop_names:
            if other_prop_name != prop_name:
                subrefs[other_prop_name] = True

        # Create ref with subrefs
        attrs[prop_name] = _make_ref(prop_name, subrefs)

    return struct(**attrs)

def _add_subref_to_ref(ref, new_subref_name):
    """Adds a new subref to an existing ref struct."""
    if type(ref) != "struct":
        return ref

    # Copy existing attributes
    attrs = {}
    for attr_name in dir(ref):
        attrs[attr_name] = getattr(ref, attr_name)

    # Add new subref if it doesn't exist
    if new_subref_name not in attrs:
        parent_name = getattr(ref, "name", None)
        if parent_name:
            attrs[new_subref_name] = _make_subref(parent_name, new_subref_name)

    return struct(**attrs)

def _add_permission_to_accessor(accessor, perm_name, perm_value, all_props, relation_types):
    """Adds a permission to the accessor struct for use by subsequent permissions."""
    # For now, make permission subrefs include all current properties
    # This is a superset that covers all possible types
    perm_subrefs = {}
    for attr_name in dir(accessor):
        perm_subrefs[attr_name] = True

    # Create ref for this permission with appropriate subrefs
    perm_ref = _make_ref(perm_name, perm_subrefs)

    # Rebuild accessor struct with new permission added
    new_attrs = {}
    for attr_name in dir(accessor):
        existing_ref = getattr(accessor, attr_name)
        # Add this permission as a subref to existing refs
        new_attrs[attr_name] = _add_subref_to_ref(existing_ref, perm_name)

    # Add the new permission itself
    new_attrs[perm_name] = perm_ref

    return struct(**new_attrs)

def permissions(resource, properties):
    # Phase 1: Filter and analyze resource properties
    relation_types = _extract_relation_types(resource)
    relation_props = {}
    for k, v in resource.items():
        if _get_property_kind(v) == "relation":
            relation_props[k] = v

    # Collect all permission names for forward references
    permission_names = list(properties.keys())

    # Phase 2: Build initial accessor struct (relations only, but with permission name subrefs)
    accessor = _build_accessor_struct_with_permissions(relation_props, relation_types, permission_names)

    # Phase 3: Process permissions sequentially, adding each to accessor before the next
    for perm_name in properties:
        factory = properties[perm_name]

        # Execute lambda with current accessor
        perm_value = factory(accessor)

        # Store in resource
        resource[perm_name] = perm_value

        # Add this permission to accessor for subsequent lambdas (so subsequent permissions can reference it)
        accessor = _add_permission_to_accessor(accessor, perm_name, perm_value, resource, relation_types)

def any(root, *refs):
    for ref in refs:
        root = root.union(ref)
    return root

def all(root, *refs):
    for ref in refs:
        root = root.intersect(ref)
    return root