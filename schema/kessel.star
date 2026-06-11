def resource(reporter, common={}, fields={}):
    return struct(kind="resource", reporter=reporter, common=common, fields=fields)

def text(minLength=None, maxLength=None, regex=None):
    return struct(kind="text", minLength=minLength, maxLength=maxLength, regex=regex)

def uuid():
    return struct(kind="uuid")

def numeric_id(min=None, max=None):
    return struct(kind="numeric_id", min=min, max=max)

def boolean():
    return struct(kind="boolean")

def date_time():
    return struct(kind="date_time")

def enum(values):
    return struct(kind="enum", values=values)

def nullable(inner):
    return struct(kind="nullable", inner=inner)

def union(left, right):
    return struct(kind="union", left=left, right=right)

def array(items):
    return struct(kind="array", items=items)

def object(properties, required=[]):
    return struct(kind="object", properties=properties, required=required)

def field(type, required=False, description=None):
    return struct(kind="field", type=type, required=required, description=description)
