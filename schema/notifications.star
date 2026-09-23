load("extensions/rbac.star", "add_v1_based_permission")

# App: Integrations - Resource: Endpoints
add_v1_based_permission("notifications", "integrations", "endpoints", "write", "integrations_endpoints_edit")
add_v1_based_permission("notifications", "integrations", "endpoints", "read", "integrations_endpoints_view")

# App: Notifications - Resource: Events
add_v1_based_permission("notifications", "notifications", "events", "read", "notifications_events_view")

# App: Notifications - Resource: Notifications
add_v1_based_permission("notifications", "notifications", "notifications", "write", "notifications_notifications_edit")
add_v1_based_permission("notifications", "notifications", "notifications", "read", "notifications_notifications_view")
