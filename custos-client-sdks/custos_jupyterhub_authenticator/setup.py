import os

from setuptools import setup, find_packages


def read(fname):
    with open(os.path.join(os.path.dirname(__file__), fname)) as f:
        return f.read()


setup(
    name='custos_jupyterhub_authenticator',
    version='1.0.0',
    packages=find_packages(),
    package_data={'': ['*.pem']},
    include_package_data=True,
    url='https://airavata.apache.org/custos/',
    license='Apache License 2.0',
    author='Custos Developers',
    author_email='dev@airavata.apache.org',
    install_requires=['oauthenticator>=14.2.0'],
    description='Apache Custos Juypterhub  Authenticator'
)
