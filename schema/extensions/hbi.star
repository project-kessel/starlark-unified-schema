_module_doc = """Wrappers for KSL extensions in the `hbi` namespace: https://github.com/project-kessel/rbac-config/blob/master/configs/prod/schemas/src/hbi.ksl"""

def _make_host_permission_reference(relation):
    def accessor(host):
        if host.kind != "ref":
            fail("Host relation is not of type 'ref' - is it a relation? (Note: sub-relations are not allowed.)")
        return host.sub(relation)
    return accessor

def expose_host_permission(v2_perm, host_perm, *legacy_reporter, reporter=None):
    """Exposes an RBAC workspace permission transitively through the host.

    The former reporter-first positional form is also accepted.

    Args:
      legacy_reporter: compatibility argument for the former reporter-first positional form.
      reporter: optional override for the file's `reporter(name)` default.
      v2_perm: the RBAC permission being exposed.
      host_perm: the name it is exposed as on the host type, used with access requests against hosts

    Returns:
      A function that accesses the workspace permission via a host relation.
    """
    if legacy_reporter:
        if len(legacy_reporter) != 1 or reporter != None:
            fail("expose_host_permission: invalid legacy reporter arguments")
        reporter, v2_perm, host_perm = v2_perm, host_perm, legacy_reporter[0]
    reporter = resolve_reporter(reporter)
    call_ksl_extension(reporter, "expose_host_permission", "hbi", v2_perm=v2_perm, host_perm=host_perm)
    return _make_host_permission_reference(host_perm)
