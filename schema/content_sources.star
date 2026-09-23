load("extensions/rbac.star", "add_v1_based_permission")

add_v1_based_permission("content_sources", "content_sources", "repositories", "read", "content_sources_repository_view")
add_v1_based_permission("content_sources", "content_sources", "repositories", "write", "content_sources_repository_edit")
add_v1_based_permission("content_sources", "content_sources", "repositories", "upload", "content_sources_repository_upload")

add_v1_based_permission("content_sources", "content_sources", "templates", "read", "content_sources_template_view")
add_v1_based_permission("content_sources", "content_sources", "templates", "write", "content_sources_template_edit")
