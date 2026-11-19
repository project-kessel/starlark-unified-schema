load("kessel.star", "resource", "uuid", "many")
load("service/reporters/features/service.star", "service")

reporter("features")

billing_account = resource(
    id_type=uuid(),
    fields={
        "services": many(service)
    })
