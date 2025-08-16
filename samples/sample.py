# [TODO] Implement login functionality
def login():
    pass


# [FIXME] Crashes when username is empty
def validate(username):
    if not username:
        raise ValueError("Username required")


# [HACK] Bypassing auth for now
def fake_auth():
    return True


# [BUG] Fails on leap years
def is_leap(year):
    return year % 4 == 0 and (year % 100 != 0 or year % 400 == 0)


# [NOTE] This is a utility function
def greet(name):
    print(f"Hello, {name}")


# [OPTIMIZE] This loop is slow for large datasets
for i in range(1000000):
    pass


# [DEPRECATED] Use new_auth() instead
def old_auth():
    return False

