_module_doc = """Wrappers for KSL extensions in the `rbac` namespace: https://github.com/project-kessel/rbac-config/blob/master/configs/prod/schemas/src/rbac.ksl"""

def _make_workspace_permission_reference(relation):
    def accessor(workspace):
        if workspace.kind != "ref":
            fail("Workspace relation is not of type 'ref' - is it a relation? (Note: sub-relations are not allowed.)")
        return workspace.sub(relation)
    return accessor

def add_v1_based_permission(app, resource, verb, v2_perm, *legacy_reporter, reporter=None):
    """Declares a v2 permission derived from an existing RBAC v1 permission.

    The former reporter-first positional form is also accepted.

    Args:
      legacy_reporter: compatibility argument for the former reporter-first positional form.
      reporter: optional override for the file's `reporter(name)` default.
      app: RBAC v1 application name, taken from the permissions json filename without the .json extension e.g. `"inventory"`.
      resource: RBAC v1 resource type, ex: `"hosts"`.
      verb: RBAC v1 verb, ex: `"read"`, `"write"`.
      v2_perm: name of the v2 permission the v1 one maps onto, ex: `"inventory_host_view"`, used with access requests against the workspace

    Returns:
      A function that accesses the permission via a workspace relation.
    """
    if legacy_reporter:
        if len(legacy_reporter) != 1 or reporter != None:
            fail("add_v1_based_permission: invalid legacy reporter arguments")
        reporter, app, resource, verb, v2_perm = app, resource, verb, v2_perm, legacy_reporter[0]
    reporter = resolve_reporter(reporter)
    call_ksl_extension(reporter, "add_v1_based_permission", "rbac", app=app, resource=resource, verb=verb, v2_perm=v2_perm)
    return _make_workspace_permission_reference(v2_perm)

def add_v1only_permission(perm, *legacy_reporter, reporter=None):
    """Declares a permission that exists in RBAC v1, used to create placeholders for permissions for unmigrated services.

    Returns nothing

    The former reporter-first positional form is also accepted.

    Args:
      legacy_reporter: compatibility argument for the former reporter-first positional form.
      reporter: optional override for the file's `reporter(name)` default.
      perm: the v1 permission name, constructed from the RBAC application, resource, and verb: {application}_{resource}_{verb}
    """
    if legacy_reporter:
        if len(legacy_reporter) != 1 or reporter != None:
            fail("add_v1only_permission: invalid legacy reporter arguments")
        reporter, perm = perm, legacy_reporter[0]
    reporter = resolve_reporter(reporter)
    call_ksl_extension(reporter, "add_v1only_permission", "rbac", perm=perm)

def add_contingent_permission(first, second, contingent, *legacy_reporter, reporter=None):
    """Declares a workspace permission that depends on two other workspace permissions, commonly used when a permission may be granted anywhere but only applies where it overlaps with another

    The former reporter-first positional form is also accepted.

    Args:
      legacy_reporter: compatibility argument for the former reporter-first positional form.
      reporter: optional override for the file's `reporter(name)` default.
      first: v2 name of the first prerequisite permission
      second: v2 name of the second prerequisite permission.
      contingent: name of the permission that applies where they overlap, used with access requests against workspaces

    Returns:
      A function that accesses the combined permission through a workspace relation
    """
    if legacy_reporter:
        if len(legacy_reporter) != 1 or reporter != None:
            fail("add_contingent_permission: invalid legacy reporter arguments")
        reporter, first, second, contingent = first, second, contingent, legacy_reporter[0]
    reporter = resolve_reporter(reporter)
    call_ksl_extension(reporter, "add_contingent_permission", "rbac", first=first, second=second, contingent=contingent)
    return _make_workspace_permission_reference(contingent)

def add_unified_permission(app, resource, verb, *legacy_reporter, reporter=None):
    """Declares a permission whose v1 and v2 names are the same.

    The v1 and v2 permission names take the form `{app}_{resource}_{verb}`, and this value is used with access requests against workspaces

    The former reporter-first positional form is also accepted.

    Args:
      legacy_reporter: compatibility argument for the former reporter-first positional form.
      reporter: optional override for the file's `reporter(name)` default.
      app: RBAC application name.
      resource: RBAC resource type.
      verb: RBAC verb.

    Returns:
      A function that accesses the combined permission through a workspace relation
    """
    if legacy_reporter:
        if len(legacy_reporter) != 1 or reporter != None:
            fail("add_unified_permission: invalid legacy reporter arguments")
        reporter, app, resource, verb = app, resource, verb, legacy_reporter[0]
    reporter = resolve_reporter(reporter)
    call_ksl_extension(reporter, "add_unified_permission", "rbac", app=app, resource=resource, verb=verb)
    return _make_workspace_permission_reference("{}_{}_{}".format(app, resource, verb))
