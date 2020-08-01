import configparser
import os

config = configparser.ConfigParser()

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
defaultSettings = os.path.join(BASE_DIR, "transport", "settings.ini")
config.read(defaultSettings)
cert_path = os.path.join(BASE_DIR,'transport','certificate.pem')


class CustosServerClientSettings(object):

    def __init__(self, configFileLocation=None):
        if configFileLocation is not None:
            config.read(configFileLocation)
        self.CUSTOS_SERVER_HOST = config.get('CustosServer', 'SERVER_HOST')
        self.CUSTOS_SERVER_PORT = config.getint('CustosServer', 'SERVER_SSL_PORT')
        self.CUSTOS_CERT_PATH = cert_path
        self.CUSTOS_CLIENT_ID = config.get('CustosServer', 'CLIENT_ID')
        self.CUSTOS_CLIENT_SEC = config.get('CustosServer', 'CLIENT_SEC')
