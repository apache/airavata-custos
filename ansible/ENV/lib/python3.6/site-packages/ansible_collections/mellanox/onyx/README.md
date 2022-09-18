# Mellanox Onyx Collection
<!--[![CI](https://zuul-ci.org/gated.svg)]-->
<!-- Add CI infornmntiom-->

The Ansible Mellanox Onyx collection includes a list of Ansible modules for managing and automating Mellanox Onyx network devices.

This collection has been tested against ONYX 3.6.8130 and above.

### Supported connections
The Mellanox Onyx collection supports ``network_cli`` connections.

## Included content

Click the ``Content`` button to see the list of content included in this collection.

## Installing this collection

You can install the Mellanox Onyx collection with the Ansible Galaxy CLI:

    ansible-galaxy collection install mellanox.onyx

You can also include it in a `requirements.yml` file and install it with `ansible-galaxy collection install -r requirements.yml`, using the format:

```yaml
---
collections:
  - name: mellanox.onyx
    version: 1.0.0
```
## Using this collection

### Using modules from the Mellanox Onyx collection in your playbooks

You can call modules by their Fully Qualified Collection Namespace (FQCN), such as `mellanox.onyx.onyx_interfaces`.

The following example task configures a network interface speed and MTU on a Mellanox Onyx network device, using the FQCN:

```yaml
---
- name: configure interface
  mellanox.onyx.onyx_interface:
      name: Eth1/2
      speed: 100G
      mtu: 512
```

Another option is to call modules by their short name if you list the `mellanox.onyx` collection in the playbook's `collections`, in the follwoing example we are creating a link aggration interface:

```yaml
---
- hosts: onyx-hosts
  gather_facts: false
  connection: network_cli

  collections:
    - mellanox.onyx

  tasks:
	- name: configure link aggregation group
	  onyx_linkagg:
	    name: Po1
	    members:
	      - Eth1/1
	      - Eth1/2
```


## Changelogs
<!--Add a link to a changelog.md file or an external docsite to cover this information. -->

## Roadmap

<!-- Optional. Include the roadmap for this collection, and the proposed release/versioning strategy so users can anticipate the upgrade/update cycle. -->

## More information

- [Ansible network resources](https://docs.ansible.com/ansible/latest/network/getting_started/network_resources.html)
- [Ansible Collection overview](https://github.com/ansible-collections/overview)
- [Ansible User guide](https://docs.ansible.com/ansible/latest/user_guide/index.html)
- [Ansible Developer guide](https://docs.ansible.com/ansible/latest/dev_guide/index.html)
- [Ansible Community code of conduct](https://docs.ansible.com/ansible/latest/community/code_of_conduct.html)

## Licensing

GNU General Public License v3.0 or later.

See [LICENSE](https://www.gnu.org/licenses/gpl-3.0.txt) to see the full text.
