def expose_host_permission(reporter, v2_perm, host_perm):
    call_ksl_extension(reporter, "expose_host_permission", "hbi", v2_perm=v2_perm, host_perm=host_perm)