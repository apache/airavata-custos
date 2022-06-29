# Airavata Custos Ansible

Ansible script to deploy Apache Airavata and PGA. 
There are ansible roles to install Airavata pre-requisites (RabbitMQ, Zookeeper, MariaDB).

## Ansible installation

Note: the following assumes a Bash shell.

1. Download and install the latest version of Python 3.6. See
   https://www.python.org/downloads/ or use your system's package manager.
2. Create a virtual environment in this directory

        cd ansible
        python3.6 -m venv ENV

3. Source the environment (you'll need to do this each time before using ansible commands)

        source ENV/bin/activate

4. Install ansible and any other dependencies.

        pip install -r requirements.txt

Now you should be ready to run `ansible-playbook` and other ansible commands.

## Useful commands

- Deploy Data Lake: `ansible-playbook -i inventories/{inventory}/ --vault-password-file=vault-password.txt custos.yml`