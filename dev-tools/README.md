# airavata-ansible

Ansible script to deploy Apache Custos

## Ansible installation

Note: the following assumes a Bash shell.

1. Download and install the latest version of Python 3.6. See
   https://www.python.org/downloads/ or use your system's package manager.
2. Create a virtual environment in this directory

        cd airavata/dev-tools/ansible
        python3.6 -m venv ENV

3. Source the environment (you'll need to do this each time before using ansible commands)

        source ENV/bin/activate

4. Install ansible and any other dependencies.

        pip install -r requirements.txt

Now you should be ready to run `ansible-playbook` and other ansible commands.

## Supported OS with versions.

- Centos 7

## Roles

- **env_setup** :- Create user and group, install oracle java 8, open firewall ports.
- **database** :- Download and install mysql(mariadb) as a service.
- **keycloak** :- Setup and deploy Keycloak Identity management server. (Note: Check roles/keycloak/README.md for details)

## Useful commands

- Deploy database: `ansible-playbook -i inventories/path/to/inventory/dir database.yml`
- Deploy Keycloak IAM server: `ansible-playbook -i inventories/path/to/inventory/dir keycloak.yml`
- ansible-playbook --vault-password-file=~/vault-password.txt -i inventories/develop keycloak.yml --tags="standalone"