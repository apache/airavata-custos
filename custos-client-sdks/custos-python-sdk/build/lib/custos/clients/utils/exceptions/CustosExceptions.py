class CustosException(BaseException):
    """ Base class for custos errors. """

    def __init__(self, *args, **kwargs):  # real signature unknown
        pass


class InvalidCredentials(CustosException):
    def __init__(self, *args, **kwargs):  # real signature unknown
        pass


class KeyAlreadyExist(CustosException):
    def __init__(self, *args, **kwargs):  # real signature unknown
        pass


class KeyDoesNotExist(CustosException):
    def __init__(self, *args, **kwargs):  # real signature unknown
        pass
