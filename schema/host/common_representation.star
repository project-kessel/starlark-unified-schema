load("kessel.star", "resource", "field", "text")

host = resource(
    common = {
        "workspace_id": field(type=text(), required=True),
    },
)
