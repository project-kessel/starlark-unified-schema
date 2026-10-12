"""Wrappers for KSL extensions in the `rbac` namespace: https://github.com/project-kessel/rbac-config/blob/master/configs/prod/schemas/src/rbac.ksl"""

def _make_workspace_permission_reference(relation):
    def accessor(workspace):
        if workspace.kind != "ref":
            fail("Workspace relation is not of type 'ref' - is it a relation? (Note: sub-relations are not allowed.)")
        return workspace.sub(relation)
    return accessor

def add_v1_based_permission(reporter, app, resource, verb, v2_perm):
    """Declares a v2 permission derived from an existing RBAC v1 permission.

    Args:
      reporter: the reporter requesting the permission.
      app: RBAC v1 application name, taken from the permissions json filename without the .json extension e.g. `"inventory"`.
      resource: RBAC v1 resource type, ex: `"hosts"`.
      verb: RBAC v1 verb, ex: `"read"`, `"write"`.
      v2_perm: name of the v2 permission the v1 one maps onto, ex: `"inventory_host_view"`, used with access requests against the workspace

    Returns:
      A function that accesses the permission via a workspace relation.
    """
    call_ksl_extension(reporter, "add_v1_based_permission", "rbac", app=app, resource=resource, verb=verb, v2_perm=v2_perm)
    return _make_workspace_permission_reference(v2_perm)

def add_v1only_permission(reporter, perm):
    """Declares a permission that exists in RBAC v1, used to create placeholders for permissions for unmigrated services.

    Returns nothing

    Args:
      reporter: the reporter requesting the permission
      perm: the v1 permission name, constructed from the RBAC application, resource, and verb: {application}_{resource}_{verb}
    """
    call_ksl_extension(reporter, "add_v1only_permission", "rbac", perm=perm)

def add_contingent_permission(reporter, first, second, contingent):
    """Declares a workspace permission that depends on two other workspace permissions, commonly used when a permission may be granted anywhere but only applies where it overlaps with another

    Args:
      reporter: the reporter requesting the permission
      first: v2 name of the first prerequisite permission
      second: v2 name of the second prerequisite permission.
      contingent: name of the permission that applies where they overlap, used with access requests against workspaces

    Returns:
      A function that accesses the combined permission through a workspace relation
    """
    call_ksl_extension(reporter, "add_contingent_permission", "rbac", first=first, second=second, contingent=contingent)
    return _make_workspace_permission_reference(contingent)

def add_unified_permission(reporter, app, resource, verb):
    """Declares a permission whose v1 and v2 names are the same.

    The v1 and v2 permission names take the form `{app}_{resource}_{verb}`, and this value is used with access requests against workspaces

    Args:
      reporter: reporter whose namespace the extension reference is written to.
      app: RBAC application name.
      resource: RBAC resource type.
      verb: RBAC verb.

    Returns:
      A function that accesses the combined permission through a workspace relation
    """
    call_ksl_extension(reporter, "add_unified_permission", "rbac", app=app, resource=resource, verb=verb)
    return _make_workspace_permission_reference("{}_{}_{}".format(app, resource, verb))