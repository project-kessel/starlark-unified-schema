def _make_host_permission_reference(relation):
    def accessor(host):
        if host.kind != "ref":
            fail("Host relation is not of type 'ref' - is it a relation? (Note: sub-relations are not allowed.)")
        return host.sub(relation)
    return accessor

def expose_host_permission(reporter, v2_perm, host_perm):
    call_ksl_extension(reporter, "expose_host_permission", "hbi", v2_perm=v2_perm, host_perm=host_perm)
    return _make_host_permission_reference(host_perm)