load("kessel.star", "resource", "uuid", "many", "at_most_one", "self", "any")

service = resource(reporter="features", id_type=uuid()) #Note: handle ancestry here and at the features workspace