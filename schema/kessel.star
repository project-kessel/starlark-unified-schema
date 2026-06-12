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
    # Just return the properties dict as-is
    # Don't set relationName here - it's only needed for permission references
    return properties

def permissions(resource, properties):
    # Create a struct wrapper where each relation becomes a ref with logic operators
    def make_accessor(d, additional_names=[], relation_types={}):
        # Collect existing relation names
        relation_names = []
        for key in d:
            relation_names.append(key)

        # Add future permission names
        for name in additional_names:
            relation_names.append(name)

        # Then create refs with sibling attributes for ALL names
        attrs = {}
        for key in d:
            ref_struct = make_ref_with_siblings(key, relation_names, relation_types)
            attrs[key] = ref_struct
        return struct(**attrs)

    def make_ref_with_siblings(name, all_relations, relation_types):
        # Create subref attributes
        nested_attrs = {}

        # Check if this relation points to another type
        target_type = relation_types.get(name)

        if target_type != None:
            # Cross-type reference: add subrefs for target type's relations
            for target_rel_name in target_type:
                nested_attrs[target_rel_name] = make_subref(name, target_rel_name)

        # Always also add same-type subrefs (for recursive permissions and sibling refs)
        for other_name in all_relations:
            if other_name != name and other_name not in nested_attrs:
                nested_attrs[other_name] = make_subref(name, other_name)

        # Create logic operator methods
        def and_method(other):
            return struct(kind="and", left=self_ref, right=other)

        def or_method(other):
            return struct(kind="or", left=self_ref, right=other)

        def unless_method(other):
            return struct(kind="unless", left=self_ref, right=other)

        # Build the complete ref struct with data, methods, and nested subrefs
        all_attrs = {
            "kind": "ref",
            "name": name,
            "intersect": and_method,
            "union": or_method,
            "exclude": unless_method,
        }

        # Merge in nested subrefs
        for k, v in nested_attrs.items():
            all_attrs[k] = v

        self_ref = struct(**all_attrs)
        return self_ref

    def make_subref(parent_name, child_name):
        # Create subref with same logic operators
        def and_method(other):
            return struct(kind="and", left=self_ref, right=other)

        def or_method(other):
            return struct(kind="or", left=self_ref, right=other)

        def unless_method(other):
            return struct(kind="unless", left=self_ref, right=other)

        self_ref = struct(
            kind="subref",
            name=parent_name,
            sub=child_name,
            intersect=and_method,
            union=or_method,
            exclude=unless_method,
        )

        return self_ref

    # Pass 1: Extract relation -> target type mappings
    relation_types = {}
    for rel_name in resource:
        rel_def = resource[rel_name]
        # Check if it's a struct (from _createRelation) with a type field
        # Use dir() to check if attribute exists, as hasattr may not work in Starlark
        if "type" in dir(rel_def):
            target_type = rel_def.type
            # Check if target_type is a dict with kind="selfType" (same-type reference)
            # For cross-type refs, target_type is a dict (the target type's relations)
            is_self_ref = False
            if type(target_type) == "dict":
                kind_val = target_type.get("kind")
                if kind_val == "selfType":
                    is_self_ref = True

            # Only store cross-type relations
            if not is_self_ref:
                relation_types[rel_name] = target_type

    # Pass 2: Collect permission names
    permission_names = []
    for name in properties:
        permission_names.append(name)

    # Pass 3: Create accessor with future permissions and type context
    resource_struct = make_accessor(resource, permission_names, relation_types)

    # Pass 4: Execute lambdas and add results
    for name in properties:
        factory = properties[name]
        prop = factory(resource_struct)

        # Store the result directly - it's already in the right format
        resource[name] = prop

def any(root, *refs):
    for ref in refs:
        root = root.union(ref)
    return root

def all(root, *refs):
    for ref in refs:
        root = root.intersect(ref)
    return root