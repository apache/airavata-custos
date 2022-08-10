# Apache Custos Ansible

Ansible scripts to configure and install Apache Custos pre-requisites and deploy the application on bare-metal web servers. The deployment process is divided into 4 steps - 
1. [Prerequisites](#prerequisites)
2. [Initial VM set up](#set-up-vms)
3. [Generate SSL certificates](#generate-ssl-certificates)
4. [Deploy application](#deploy-application)

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
git clone https://github.com/abhinav7sinha/airavata-custos.git
cd airavata-custos
git switch ansible-baremetal
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

Open the file `test/group_vars/all/vars.yml`. Look for `keycloak_domain`, `hashicorp_domain` and `custos_domain` variables. Set these variables with their respective domain names. For ex:
```yml
⋮
# Domain names for VMs
keycloak_domain: my-keycloak-domain.com
hashicorp_domain: my-hashicorp-domain.com
custos_domain: my-custos-domain.com
⋮
```
Ansible connects to these VMs using `ssh`. To allow ansible to do this, you will also need to provide it with ssh passwords for these VMs. To maintain security, these passwords cannot be exposed in plaintext. You can make use of ansible's `ansible-vault` command to encrypt these password strings.

Use the following command to encrypt any string:
```bash
ansible-vault encrypt_string "<mysecurepassword>" --ask-vault-pass
```
Ansible will ask you to set up a new vault password everytime you run this command. Make sure to use the same password everytime.

In the file `test/group_vars/all/vars.yml`. Look for `keycloak_ssh_password`, `hashicorp_ssh_password` and `custos_ssh_password` variables. Set these variables with their respective encrypted password strings. For ex:
```yml
⋮
keycloak_ssh_password: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  36363039363662616335356566613665346161396138303131616334616666623361633765356434
  3234326462633763636...43161313064
hashicorp_ssh_password: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  62633531346536643234303565373639626464326135653732323266396335353166346132383230
  3333316665623861326...462353738656662
custos_ssh_password: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  643562326362326461334303565373639626464326135653732323266396335353166346132383230
  3333316665623861326...065626463633330
⋮
```
Now that you've configured the ssh user and password details for your VMs, we can install nginx on each of these servers at once by running the following command:
```bash
ansible-playbook -i inventories/test/ env_setup.yml --ask-vault-pass
```
Enter the Vault Password (that you set up while encrypting your ssh password strings) when prompted by ansible.

## Generate SSL certificates<a name="generate-ssl-certificates"/>
The `env_setup.yml` playbook installed `certbot` along with `nginx` in your VMs. **Certbot** helps manage letsencrypt SSL certificates. You need to generate SSL certs for each of your 3 VMs.

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
openssl pkcs12 -export -out certs/vault-client-truststore.pkcs12 -inkey /etc/letsencrypt/live/<my-hashicorp-domain.com>/privkey.pem -in /etc/letsencrypt/live/<my-hashicorp-domain.com>/fullchain.pem
```
In the keycloak VM:
```bash
openssl pkcs12 -export -out certs/keycloak-client-truststore.pkcs12 -inkey /etc/letsencrypt/live/<my-keycloak-domain.com>/privkey.pem -in /etc/letsencrypt/live/<my-keycloak-domain.com>/fullchain.pem
```
Now, you need to add the pkcs12 passwords in the ansible vars.

Encrypt the passwords using `ansible-vault` just like you did for the ssh passwords. Make sure you still use the same vault password as before. In the file `test/group_vars/all/vars.yml`. Look for `hashicorp_pkcs12_passphrase` and `keycloak_pkcs12_passphrase` variables. Set these variables with their respective encrypted password strings. For ex:
```yml
⋮
# pkcs12 passwords
hashicorp_pkcs12_passphrase: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  65646665313232613265363038633662393438376265663363333434303361656565653763313539
  393735653837326...537
keycloak_pkcs12_passphrase: !vault |
  $ANSIBLE_VAULT;1.1;AES256
  63303962613033353165646639633338306538666233363735393266386432376663656536633663
  62313163333631...831303961
⋮
```

## Deploy application<a name="deploy-application"/>
You're just one step away from deploying Custos. The following command runs the ansible playbook `custos.yml` which has roles to install all dependencies and deploy Custos on the given VM.
```bash
ansible-playbook -i inventories/test/ custos.yml --ask-vault-pass
```
Check if custos core and integration services are up and running using the following commands:
```bash
systemctl status intcustos
```
```bash
systemctl status corecustos
```

## Useful commands
- Deploy Custos with verbose option for debug messages:
```bash
ansible-playbook -i inventories/{inventory}/ custos.yml --ask-vault-pass -vvv
```
Adding multiple -v will increase the verbosity, the builtin plugins currently evaluate up to -vvvvvv. A reasonable level to start is -vvv, connection debugging might require -vvvv.
- Decrypt ansible-vault strings
```bash
ansible all -i inventories/test/ -e '@inventories/test/group_vars/all/vars.yml' --ask-vault-pass -m debug -a 'var=secret_string'
```

