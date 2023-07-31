# Apache Custos Ansible

Ansible scripts to configure and install Apache Custos pre-requisites and deploy the application on bare-metal web servers. The deployment process is divided into 4 steps - 
1. [Prerequisites](#prerequisites)
2. [Initial VM set up](#set-up-vms)
3. [Generate SSL certificates](#generate-ssl-certificates)
4. [Deploy application](#deploy-application)
5. [Deploy application with data migration](#data-migration)

## Prerequisites<a name="prerequisites"/>
Make sure you have installed all of the following prerequisites on your development machine:
* Git - [Download & Install Git](https://git-scm.com/downloads). OSX and Linux machines typically have this already installed.
* Python3 - [Download and Install Python3](https://www.python.org/downloads/). You can also use your system's package manager to install the latest stable version of python3.

Note: the following assumes a Bash shell.

Verify if git was installed successfully.
```bash
git --version
```
Verify if python3 was installed successfully.
```bash
python3 --version
```
Clone Custos git repo and switch branch.
```bash
git clone https://github.com/apache/airavata-custos.git
cd airavata-custos
git switch develop
```
Create a virtual environment in the ansible directory
```bash
cd ansible
python3 -m venv ENV
```
Activate the virtual environment. (You'll need to do this each time before using ansible commands)
```bash
source ENV/bin/activate
```
The requirements.txt file contains ansible and any other dependencies. Install it by running the following command:
```bash
pip install -r requirements.txt
```
## Set up VMs<a name="set-up-vms"/>
You will need 3 Linux VMs (Preferably Ubuntu 20.04) for Custos Deployment. Make sure you have registered domain names for each of them.
* _VM #1_: for hosting **Keycloak**
* _VM #2_: for hosting **Hashicorp Vault** and **Consul**
* _VM #3_: for hosting **Custos Spring Boot application**.

### Configure Ansible vars with VM details
Now, you need to configure Ansible vars with the domain names of these VMs so that ansible knows where to install custos dependencies.

Open the file `test/hosts.yml`. Set `ansible_host` for `custos`, `keycloak` and `hashicorp` with their respective domain names. For ex:

```yml
all:
  hosts:
    custos:
      ansible_host: my-custos-domain.com
      ⋮
    keycloak:
      ansible_host: my-keycloak-domain.com
      ⋮
    hashicorp:
      ansible_host: my-hashicorp-domain.com
      ⋮
```
Ansible connects to these VMs using `ssh`. To allow ansible to do this, you will also need to provide it with ssh passwords for these VMs. To maintain security, these passwords cannot be exposed in plaintext. We use ansible's `ansible-vault` command to encrypt these password strings.

Open the file `test/group_vars/all/vault.yml`. You will need to decrypt the file to access it (it's been encrypted using ansible-vault for security).
Make sure you put the password text in this file - `ansible/vault_pass`. You can reach out to the dev team to get the password for test env.
Here's the command:
```bash
ansible-vault decrypt inventories/test/group_vars/all/vault.yml
```

Anytime you change the `vault.yml` file, don't forget to encrypt it back again.
Here's the command to encrypt:
```bash
ansible-vault encrypt inventories/test/group_vars/all/vault.yml
```

Now that you've configured the ssh user and password details for your VMs, we can install Java 17 on each of these servers at once by running the following command:
```bash
ansible-playbook -i inventories/test/ custos.yml --tags env_setup
```

## Generate SSL certificates<a name="generate-ssl-certificates"/>
The `env_setup.yml` playbook installed `certbot` along with `nginx` in your VMs. 

Now **Certbot** will help us manage letsencrypt SSL certificates. You need to generate SSL certs for each of your 3 VMs.

### Generate letsencrypt certificates
SSH into your VMs with your user:
```bash
ssh <myuser>@<my-domain.com>
```
Run the following command:
```bash
sudo certbot --nginx -d <my-domain.com> -m <myemail@gmail.com> --agree-tos --no-eff-email --redirect
```
Certbot generates a cert file (`fullchain.pem`) and a keyfile (`privkey.pem`) in the following location in your VM - `/etc/letsencrypt/live/<my-domain.com>/`

### Generate pkcs12 certs from pem files
Run the following commands to generate pkcs12 truststore files from you cert and key files. You will be asked to create a password for the pkcs12 file, keep the password in a safe location.

In the hashicorp VM:
```bash
sudo openssl pkcs12 -export -out certs/vault-client-truststore.pkcs12 -inkey /etc/letsencrypt/live/<my-hashicorp-domain.com>/privkey.pem -in /etc/letsencrypt/live/<my-hashicorp-domain.com>/fullchain.pem
```
In the keycloak VM:
```bash
sudo openssl pkcs12 -export -out certs/keycloak-client-truststore.pkcs12 -inkey /etc/letsencrypt/live/<my-keycloak-domain.com>/privkey.pem -in /etc/letsencrypt/live/<my-keycloak-domain.com>/fullchain.pem
```
You should see the respective pkcs12 file in the `~/certs` folder.

Now, you need to add the pkcs12 passwords in the ansible vars.

In the file `test/group_vars/all/vault.yml`. Look for `hashicorp_pkcs12_passphrase` and `keycloak_pkcs12_passphrase` variables. Set these variables with their respective password strings. For ex:
```yml
⋮
# pkcs12 passwords
hashicorp_pkcs12_passphrase: XXX...XXX
keycloak_pkcs12_passphrase: YYY...YYY
⋮
```

## Deploy application<a name="deploy-application"/>

You're just one step away from deploying Custos. The following command runs the ansible playbook `custos.yml` which has roles to install all dependencies and deploy Custos on the given VM.
```bash
ansible-playbook -i inventories/test/ custos.yml
```

This will install HashiCorp Vault, Keycloak and Custos Services. Please edit the ```custos.yml``` 
according to deploying services. By default vault, Keycloak and Custos services will be deployed.

### Keycloak Installation Steps
Run ansible  with host keycloak. Successfull run should enable keycloak service on host machine.Please check
```bash
systemctl status keycloak
```

You have to manually run  ```add-user-keycloak.sh``` to add initial admin user.Once you ran it should be able to
login to the keycloak portal via URL  ```https://<keycloak.host>/auth/```.

### Vault Installation Steps

Run ansible  with host vault. Successfull run should enable vault service on host machine.Please check
```bash
systemctl status vault
```
Configure vault.yml following properties based on above properties.

```bash
vault_token, custos_core_iam_server_admin_username,
custos_core_iam_server_admin_password, custos_core_ciLogon_admin_client_id,
custos_core_ciLogon_admin_client_secret
```
And rest of the properties accordingly.

Login to vault UI URL and initialize the vault.URL should be  ```https://<vault.host>:8200/ui```.

### Custos Service Installation
Run ansible  with custos host. Check if custos core and integration services are up and running using the following commands:
```bash
systemctl status intcustos
```
```bash
systemctl status corecustos
```
Note: You have it regenerate ```keycloak-client-truststore.pkcs12``` and ```vault-client-truststore.pkcs12``` If those are expired.


## Deploy application with data migration<a name="data-migration"/>
When deploying Custos, you may want to import data from your old servers to the new ones. We can make use of ansible to automate this data migration.

You need to configure ansible vars with the domain names of the old VMS.

Open the file `test/hosts.yml`. Set ansible_host and ansible_user for `old_custos`, `old_keycloak` and `old_hashicorp` with their respective domain names and ssh users. For ex:
```yml
    ⋮
    old_custos:
      ansible_host: oldcustos.org
      ansible_user: test_user
    old_keycloak:
      ansible_host: oldkeycloak.org
      ansible_user: test_user
    old_hashicorp:
      ansible_host: oldvault.org
      ansible_user: test_user
    ⋮
```
Make sure that passwordless ssh authentication is set up from your master node to the old servers.

Now Generate backups in your old servers. By default, ansible will load the backups from the locations specified in the file: `ansible/roles/migrate_db/tasks/main.yml`

To deploy Custos with data migration, run the following command:
```bash
ansible-playbook -i inventories/test/ custos.yml --tags migrate_db,all
```
## Useful commands
- Deploy Custos with verbose option for debug messages:
```bash
ansible-playbook -i inventories/{inventory}/ custos.yml -vvv
```
Adding multiple -v will increase the verbosity, the builtin plugins currently evaluate up to -vvvvvv. A reasonable level to start is -vvv, connection debugging might require -vvvv.
- Decrypt ansible-vault strings
```bash
ansible all -i inventories/test/ -e '@inventories/test/group_vars/all/vars.yml' -m debug -a 'var=secret_string'
```
- Encrypt a string using `ansible-vault`
```bash
ansible-vault encrypt_string "MySecureString"
```
