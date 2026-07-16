# Composite Workspace Schema — Solution 3: Common Workspace with Prefixed Relations
#
# Same traversal and answer as Solution 2, but instead of adding Features relations
# directly to rbac/workspace, all services contribute to a single common/workspace
# type with prefixed relation names to avoid collisions.
#
# Key differences from alternative-features-schema.star (Solution 2):
#   - Workspace type is common/workspace, not rbac/workspace
#   - Each service's relations are prefixed: rbac_*, features_*, etc.
#   - Multiple services can contribute workspace config without modifying each other
#   - Host points to common/workspace and accesses any service's permissions
#
# Check: zed permission check hbi/host:123 enabled features/service:RHEL
# Path:  host → common/workspace → parent* → features_available_services → service:RHEL

load("kessel.star", "resource", "uuid", "many", "at_most_one", "self", "one", "any")

# --- Service ---
# No relations. Service is just an identity.

service = resource(reporter="features", id_type=uuid(), fields={})

# --- Billing Account ---
# Same as Solution 2 — billing account points to services and has a parent hierarchy.

billing_account = resource(reporter="features", id_type=uuid(),
fields={
    "direct_services": many(service),
    "parent": at_most_one(self()),
},
permissions={
    "services": lambda b: b.direct_services.union(b.parent.services),
})

# --- Common Workspace ---
# Single composite type that merges all services' workspace-specific relations.
# Each service's relations are prefixed to avoid name collisions.
#
# This replaces both rbac/workspace and any per-service workspace types.
# All services report their workspace-specific data to this one type.
#
# In SpiceDB, this compiles to:
#
#   definition common/workspace {
#       relation t_parent: common/workspace
#       relation t_rbac_binding: rbac/role_binding
#       relation t_features_direct_billing_accounts: features/billing_account
#       relation t_features_direct_service_preference: features/service
#       permission features_available_services = features_paid_services & features_desired_services
#       ...
#   }

workspace = resource(reporter="common", id_type=uuid(),
fields={
    "parent": at_most_one(self()),

    # RBAC-specific (would be prefixed rbac_ in SpiceDB)
    "rbac_binding": many(self()),  # placeholder for role_binding type

    # Features-specific (would be prefixed features_ in SpiceDB)
    "features_direct_billing_accounts": many(billing_account),
    "features_direct_service_preference": many(service),
},
permissions={
    # Features computed permissions
    "features_paid_services": lambda w: w.features_direct_billing_accounts.services.union(
        w.parent.features_paid_services,
    ),
    "features_desired_services": lambda w: w.features_direct_service_preference.union(
        w.parent.features_desired_services,
    ),
    "features_available_services": lambda w: w.features_paid_services.intersect(
        w.features_desired_services,
    ),

    # RBAC computed permissions (example, follows same parent-up pattern)
    # "rbac_inventory_host_view": lambda w: w.rbac_binding.inventory_host_view.union(
    #     w.parent.rbac_inventory_host_view,
    # ),
})

# --- Host ---
# Points to common/workspace. Can access both RBAC and Features permissions
# through the same workspace reference.
#
# In SpiceDB:
#
#   definition hbi/host {
#       relation workspace: common/workspace
#       permission enabled = workspace->features_available_services
#       permission view = workspace->rbac_inventory_host_view
#   }

host = resource(reporter="hbi", id_type=uuid(),
common={
    "workspace_id": one(workspace),
},
fields={},
permissions={
    "enabled": lambda h: h.workspace_id.features_available_services,
    # "view": lambda h: h.workspace_id.rbac_inventory_host_view,
})
