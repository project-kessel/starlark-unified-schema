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
    # Note: 'and' and 'or' are reserved keywords, so we use 'and_' and 'or_'
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
    def make_accessor(d, additional_names=[]):
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
            ref_struct = make_ref_with_siblings(key, relation_names)
            attrs[key] = ref_struct
        return struct(**attrs)

    def make_ref_with_siblings(name, all_relations):
        # Create subref attributes for all OTHER relations
        nested_attrs = {}
        for other_name in all_relations:
            if other_name != name:
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

    # Pass 1: Collect permission names
    permission_names = []
    for name in properties:
        permission_names.append(name)

    # Pass 2: Create accessor with future permissions included
    resource_struct = make_accessor(resource, permission_names)

    # Pass 3: Execute lambdas and add results
    for name in properties:
        factory = properties[name]
        prop = factory(resource_struct)

        # Store the result directly - it's already in the right format
        resource[name] = prop
