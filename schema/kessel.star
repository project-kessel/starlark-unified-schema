def _(left, operator, right):
    return struct(kind=operator, left=left, right=right)

def self():
    return struct(kind="selfType")

def _createRelation(kind, type):
    def sub(name):
        return struct() #Create a subrelation expression from the name and type
    
    return struct(kind=kind, type=type, sub=sub)

def atMostOne(type):
    return _createRelation("atMostOne", type)

def one(type):
    return _createRelation("exactlyOne", type)

def atLeastOne(type):
    return _createRelation("atLeastOne", type)

def many(type):
    return _createRelation("many", type)

def all(type):
    return _createRelation("boolean", type)