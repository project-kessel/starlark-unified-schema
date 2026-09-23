def add_v1_based_permission(reporter, app, resource, verb, v2_perm):
    call_ksl_extension(reporter, "add_v1_based_permission", "rbac", app=app, resource=resource, verb=verb, v2_perm=v2_perm)

def add_v1only_permission(reporter, perm):
    call_ksl_extension(reporter, "add_v1only_permission", "rbac", perm=perm)

def add_contingent_permission(reporter, first, second, contingent):
    call_ksl_extension(reporter, "add_contingent_permission", "rbac", first=first, second=second, contingent=contingent)

def add_unified_permission(reporter, app, resource, verb):
    call_ksl_extension(reporter, "add_unified_permission", "rbac", app=app, resource=resource, verb=verb)