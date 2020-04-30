from base64 import b64encode


def get_token(custos_settings):
    tokenStr = custos_settings.CUSTOS_CLIENT_ID + ":" + custos_settings.CUSTOS_CLIENT_SEC
    tokenByte = tokenStr.encode('utf-8')
    encodedBytes = b64encode(tokenByte)
    return encodedBytes.decode('utf-8')
