load("kessel.star", "resource", "uuid", "many")
load("service/reporters/features/service.star", "service")

billing_account = resource(reporter="features", 
id_type=uuid(), 
fields={
    "services": many(service)
})
