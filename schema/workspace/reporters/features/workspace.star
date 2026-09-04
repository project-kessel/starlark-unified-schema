load("kessel.star", "resource", "at_most_one", "many", "wildcard", "any", "uuid")
load("billing_account/reporters/features/billing_account.star", "billing_account")
load("service/reporters/features/service.star", "service")
load("workspace/reporters/rbac/workspace.star", rbac_workspace="workspace")

workspace = resource("features", extends=rbac_workspace, fields={
    "direct_billing_account": at_most_one(billing_account),
    "direct_service_preferences": many(service),
    "desire_all_services": wildcard(service),
    "inherit_desired_services": wildcard(service),
    "inherit_paid_services": wildcard(service)
}, permissions={
    "_paid_services": lambda w: any(
        w.direct_billing_account.services,
        w.parent._paid_services.intersect(w.inherit_paid_services)
    ),
    "_desired_services": lambda w: any(
        w.direct_service_preferences,
        w.desire_all_services,
        w.parent._desired_services.intersect(w.inherit_desired_services)
    ),
    "enabled_services": lambda w: w._paid_services.intersect(w._desired_services)
})
