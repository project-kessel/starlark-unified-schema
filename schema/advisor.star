load("extensions/rbac.star", "add_v1_based_permission", "add_contingent_permission")
load("extensions/hbi.star", "expose_host_permission")

add_v1_based_permission("advisor", "advisor", "disable_recommendations", "read", "advisor_disable_recommendations_view")
add_v1_based_permission("advisor", "advisor", "disable_recommendations", "write", "advisor_disable_recommendations_edit")

add_v1_based_permission("advisor", "advisor", "weekly_email", "read", "advisor_weekly_email_view")

add_v1_based_permission("advisor", "advisor", "weekly_report", "read", "advisor_weekly_report_view")
add_v1_based_permission("advisor", "advisor", "weekly_report", "write", "advisor_weekly_report_edit")

add_v1_based_permission("advisor", "advisor", "weekly_report_auto_subscribe", "read", "advisor_weekly_report_auto_subscribe_view")
add_v1_based_permission("advisor", "advisor", "weekly_report_auto_subscribe", "write", "advisor_weekly_report_auto_subscribe_edit")

add_v1_based_permission("advisor", "advisor", "recommendation_results", "read", "advisor_recommendation_results_view_assigned")
add_contingent_permission("advisor", "inventory_host_view", "advisor_recommendation_results_view_assigned", "advisor_recommendation_results_view")
expose_host_permission("advisor", "advisor_recommendation_results_view", "advisor_recommendation_results_view")

add_v1_based_permission("advisor", "advisor", "recommendation_results", "write", "advisor_recommendation_results_edit_assigned")
add_contingent_permission("advisor", "inventory_host_view", "advisor_recommendation_results_edit_assigned", "advisor_recommendation_results_edit")
expose_host_permission("advisor", "advisor_recommendation_results_edit", "advisor_recommendation_results_edit")

add_v1_based_permission("advisor", "advisor", "exports", "read", "advisor_exports_view")
