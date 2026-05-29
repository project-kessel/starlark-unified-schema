load("kessel.star", "field", "uuid", "text", "nullable", "union")

resource("host",
    common = {
        "workspace_id": field(type=text(), required=True),
    },
    reporters = {
        "hbi": {
            "satellite_id": field(type=nullable(union(uuid(), text(regex="^\\d{10}$")))),
            "subscription_manager_id": field(type=nullable(uuid())),
            "insights_id": field(type=nullable(uuid())),
            "ansible_host": field(type=nullable(text(maxLength=255))),
        },
    },
)
