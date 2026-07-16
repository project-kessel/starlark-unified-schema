# Alternative Features Schema — Solution 2: Flip Features Config onto Workspace
#
# Instead of service → workspace (feature-simple.star), this puts billing accounts
# and service preferences ON the workspace, using parent-up inheritance.
#
# Key differences from feature-simple.star:
#   - No children/descendants on workspace (no downward traversal)
#   - Workspace owns the billing account and service preference relations
#   - Inheritance uses parent-up (O(log n)) instead of descendant expansion (O(n))
#   - Check starts at the host, not the service:
#       zed permission check hbi/host:123 enabled features/service:RHEL
#
# Comparison:
#   feature-simple.star:        service → workspace (+ children/descendants)
#   this file:                  host → workspace → service (parent-up only)

load("kessel.star", "resource", "uuid", "many", "at_most_one", "self", "one")

# --- Service ---
# No relations pointing to workspaces. Service is a simple identity.

service = resource(reporter="features", id_type=uuid(), fields={})

# --- Billing Account ---
# Billing account points to the services it pays for.
# Direction: billing_account → service (billing account is resource, service is subject)

billing_account = resource(reporter="features", id_type=uuid(),
fields={
    "direct_services": many(service),
    "parent": at_most_one(self()),
},
permissions={
    "services": lambda b: b.direct_services.union(b.parent.services),
})

# --- Workspace ---
# Workspace points to billing accounts and service preferences.
# Direction: workspace → billing_account, workspace → service
# Inheritance: parent-up (same pattern as existing RBAC permissions)

workspace = resource(reporter="rbac", id_type=uuid(),
fields={
    "parent": at_most_one(self()),
    "direct_billing_accounts": many(billing_account),
    "direct_service_preference": many(service),
},
permissions={
    "_paid_services": lambda w: w.direct_billing_accounts.services.union(w.parent._paid_services),
    "_desired_services": lambda w: w.direct_service_preference.union(w.parent._desired_services),
    "available_services": lambda w: w._paid_services.intersect(w._desired_services),
})

# --- Host ---
# Host points to workspace (existing). Adds permission to traverse to available services.
#
# Check: zed permission check hbi/host:123 enabled features/service:RHEL
# Path:  host → workspace → parent* → available_services → service:RHEL

host = resource(reporter="hbi", id_type=uuid(),
common={
    "workspace_id": one(workspace),
},
fields={},
permissions={
    "enabled": lambda h: h.workspace_id.available_services,
})
