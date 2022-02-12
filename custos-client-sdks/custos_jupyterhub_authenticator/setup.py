import os

from setuptools import setup, find_packages


def read(fname):
    with open(os.path.join(os.path.dirname(__file__), fname)) as f:
        return f.read()


from pathlib import Path

this_directory = Path(__file__).parent
long_description = (this_directory / "README.md").read_text()

setup(
    name='custos_jupyterhub_authenticator',
    long_description_content_type="text/markdown",
    version='1.0.2',
    packages=find_packages(),
    package_data={'': ['*.pem']},
    include_package_data=True,
    url='https://github.com/apache/airavata-custos/tree/develop/custos-client-sdks/custos_jupyterhub_authenticator',
    license='Apache License 2.0',
    author='Custos Developers',
    author_email='dev@airavata.apache.org',
    install_requires=['oauthenticator>=14.2.0'],
    description='Apache Custos Jupyterhub  Authenticator',
    long_description=long_description,
)
