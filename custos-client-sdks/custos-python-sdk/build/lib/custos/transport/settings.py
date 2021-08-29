import configparser
import os

config = configparser.ConfigParser()

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
defaultSettings = os.path.join(BASE_DIR, 'transport', 'settings.ini')
config.read(defaultSettings)
cert_path = os.path.join(BASE_DIR, 'transport', 'certificate.pem')


class CustosServerClientSettings(object):

    def __init__(self, configuration_file_location=None, custos_host=None,
                 custos_port=None, custos_client_id=None, custos_client_sec=None):
        if configuration_file_location is not None:
            config.read(configuration_file_location)
        self.CUSTOS_CERT_PATH = cert_path

        if custos_host is not None:
            self.CUSTOS_SERVER_HOST = custos_host
        else:
            self.CUSTOS_SERVER_HOST = config.get('CustosServer', 'SERVER_HOST')
        if custos_port is not None:
            self.CUSTOS_SERVER_PORT = custos_port
        else:
            self.CUSTOS_SERVER_PORT = config.getint('CustosServer', 'SERVER_SSL_PORT')
        if custos_client_id is not None:
            self.CUSTOS_CLIENT_ID = custos_client_id
        else:
            self.CUSTOS_CLIENT_ID = config.get('CustosServer', 'CLIENT_ID')

        if custos_client_sec is not None:
            self.CUSTOS_CLIENT_SEC = custos_client_sec
        else:
            self.CUSTOS_CLIENT_SEC = config.get('CustosServer', 'CLIENT_SEC')
