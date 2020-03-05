import configparser
import os

config = configparser.ConfigParser()

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
defaultSettings = os.path.join(BASE_DIR, "transport", "settings.ini")
config.read(defaultSettings)


class CustosServerClientSettings(object):

    def __init__(self, configFileLocation=None):
        if configFileLocation is not None:
            config.read(configFileLocation)
        self.CUSTOS_SERVER_HOST = config.get('CustosServer', 'SERVER_HOST')
        self.CUSTOS_SERVER_PORT = config.getint('CustosServer', 'SERVER_SSL_PORT')
        self.CUSTOS_CERT_PATH = config.get('CustosServer', 'CERTIFICATE_FILE_PATH')
