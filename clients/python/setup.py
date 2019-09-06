
from setuptools import setup, find_packages

setup_requirements = ['pytest-runner', ]

test_requirements = ['pytest', ]

setup(
    name='airavata_custos',
    version="0.0.1",
    description='Apache Airavata Custos Python API',
    author='Airavata Custos Developers',
    author_email='custos@airavata.apache.org',
    packages=find_packages(include=['airavata_custos']),
    license='Apache License 2.0',
    setup_requires=setup_requirements,
    test_suite='tests',
    tests_require=test_requirements,
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.6'
    ]
)
