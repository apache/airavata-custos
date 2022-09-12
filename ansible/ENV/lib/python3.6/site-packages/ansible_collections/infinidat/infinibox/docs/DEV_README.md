# Testing of Infinidat's Ansible Modules

## Conventions
We use "yaml" as the extention for configuration files and "yml" for playbooks and role main.yml files.

## Set up
- Clone ansible-infinidat-collection from one of:
    - https://github.com/infinidat/ansible-infinidat-collection (external)
    - https://git.infinidat.com/PSUS/ansible-infinidat-collection (internal)
- Create a Python virtualenv:
    - `cd ansible-infinidat-collection`
    - `python3 -m venv venv`
    - `source venv/bin/activate`
    - `sudo apt install libpython3.8-dev libffi-dev`
    - `pip install -r requirements.txt`

## Creating playbook extra-var files
Extra-var files define some Ansible variable necessary for running playbooks.  There is an example ./ibox_vars/iboxNNNN_example.yaml.  The auto_prefix value is used in the test playbooks within the names of most resources created on the Infinibox.

Example:
```
    auto_prefix: "PSUS_ANSIBLE_"
    user: "user"
    password: "passwd"
    system: "ibox2233"
```

Extra-var files containing secrets should be encrypted using ansible-vault:
- `ansible-vault [encrypt,view,edit] <extra-var file>`

## Creating a collection
Ansible collections require Ansible 2.9+. Complete instructions for creating and using collections is available from Ansible at https://docs.ansible.com/ansible/latest/cli/ansible-galaxy.html.

A collection is a tarball with a specified set of files and directories that meet Ansible's collection requirements.

## Creating resources
- `cd ./projects/playbooks`
- The ask-vault-pass option is required if the extra-var file is encrypted.
- `../venv/bin/ansible-playbook --extra-vars "@../ibox_vars/iboxNNNN.yaml" --ask-vault-pass test_create_resources.yml`
    - Idempotency: Running playbooks repeatedly work correctly with all tasks showing "ok" (green).

## Removing resources
- `cd ./projects/playbooks`
- `../venv/bin/ansible-playbook --extra-vars "@../ibox_vars/iboxNNNN.yaml" --ask-vault-pass test_remove_resources.yml`
    - Running this repeatedly should work correctly with all tasks showing "ok" (green).

## Hacking
Hacking lets one execute a module as a normal python script without Ansible. This allows use of print() and pdb for debugging. When executing modules using ansible-playbook, within a task, print statement output from the module are not displayed.

Modules expect a JSON data file to be provided to them.  This represents the equivalent JSON that Ansible would provide the module if the module is called from an Ansible task.   The modules also return results to stdout in JSON form.

### Creating hacking JSON files
Keys within JSON files must define values that match the requirements of the module to be tested. They are exactly equivalent to the fields defined in a task using the same module.

You may find the requirements of a module using `ansible-doc`.
- Example:
    - `ansible-doc infini_vol`
        - See the `library` variable in the ansible.cfg file.
    - `ansible-doc --module-path=~/ansible/lib/ansible/modules/storage/infinidat infini_export_client`
        - This less useful example shows using the module-path option if not defined in your ansible.cfg file. In this case it is providing documentation from a github clone of the Ansible project and the older Infinidat modules Ansible comes with.

#### Example JSON files:
- `test_logout.json`:
  ```
  {
      "ANSIBLE_MODULE_ARGS": {
          "name": "PSUS_ANSIBLE_logout_pool",
          "size": "1TB",
          "vsize": "1TB",
          "state": "present",
          "user": "user",
          "password": "passwd",
          "system": "ibox1339.lab.gdc.il.infinidat.com"
      }
  }
  ```
- `test_export_fs.json`:
  ```
  {
      "ANSIBLE_MODULE_ARGS": {
          "name": "/PSUS_ANSIBLE_export",
          "filesystem": "PSUS_ANSIBLE_fs",
          "client_list": [
              {
                  "client": "*",
                  "access": "RO",
                  "no_root_squash": true
              }
          ],
          "state": "present",
          "user": "user",
          "password": "passwd",
          "system": "ibox1339.lab.gdc.il.infinidat.com"
      }
  }
  ```
- `test_cluster.json`:
  ```
  {
      "ANSIBLE_MODULE_ARGS": {
          "name": "PSUS_ANSIBLE_cluster",
          "cluster_hosts": [
              {
                "host_name": "PSUS_ANSIBLE_host",
                "host_cluster_state": "present"
              }
          ],
          "state": "present",
          "user": "admin",
          "password": "passwd",
          "system": "ibox1339.lab.gdc.il.infinidat.com"
      }
  }
  ```

### Executing modules without playbooks (hacking)
To run Ansible modules directly with Python and without using ansible-playbook, extra steps are required.  These are documented by [Ansible module development](https://docs.ansible.com/ansible/2.9/dev_guide/developing_modules_general.html).

Once your hacking environment is set up, you must copy ansible-infinidat-collection modules, etc. into the appropriate places within your Ansible clone.  The virtualenv described above should be used.  Source `hacking/env-setup` after activating the virtualenv.
```
export ansible_clone="<path to Ansible clone>"
cd ansible-infinidat-collection/
cp plugins/modules/infini*.py           "$ansible_clone/lib/ansible/modules/storage/infinidat/"
cp plugins/module_utils/infinibox.py    "$ansible_clone/lib/ansible/module_utils/"
cp plugins/doc_fragments/infinibox.py   "$ansible_clone/lib/ansible/plugins/doc_fragments/"
```

`bin/install_modules_for_hacking.sh` may be used to copy the files described above. If executed from anywhere with in the ansible-infinidat-collection working copy, it will find the proper files and copy them into the specified Ansible source clone. The existing files in the clone will be overwritten.
```
./bin/install_modules_for_hacking.sh <path to clone of Ansible's src>
```

Run/hack a module:
```
export test_jsons="<path to test JSON files>"
python -m ansible.modules.storage.infinidat.infini_host "$test_jsons/test_host.json"
python -m ansible.modules.storage.infinidat.infini_host "$test_jsons/test_host.json" 2>&1 | grep -v Insecure
python -m ansible.modules.storage.infinidat.infini_host "$test_jsons/test_host.json" 2>&1 | grep -v Insecure | jq --sort-keys '.'
```

### Set colors for those who are color challenged (optional)
Use of jq and pygmentize is not required, but we've found it useful.  Use pygmentize to colorize JSON data using the autumn theme.  Pymentize may be installed via pip.

Use jq to pretty print JSON. See https://stedolan.github.io/jq/
- -c: Optionally, use compact output. This displays JSON in a much shorter form, but may be harder to read.
- --sort-keys: Sort keys.
- The period is the simplest jq filter.  It is a null filter showing all input.

```
pyg_style="autumn"
jqpyg="pygmentize -O style=$pyg_style -l json"
python -m ansible.modules.storage.infinidat.infini_port test_ports_port.json 2>&1 | grep -v Insecure  | jq --sort-keys -c . | eval $jqpyg
```
