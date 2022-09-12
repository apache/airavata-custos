# A Makefile for creating, running and testing Infindat's Ansible collection.

### Dependencies ###
# - jq: https://stedolan.github.io/jq/
# - spruce: https://github.com/geofffranks/spruce

### environment ###
# Include an env file with secrets.  This exposes the secrets
# as envvars only for the life of make.  It does not
# pollute the environment persistently.
# Format:
# API_KEY=someAnsibleGalaxyApiKey
# The key only needs to be valid to use target galaxy-colletion-publish.
_env = ~/.ssh/ansible-galaxy.sh
include $(_env)
export $(shell sed 's/=.*//' $(_env))

### Vars ###
_version            = $(shell spruce json galaxy.yml | jq '.version'   | sed 's?"??g')
_namespace          = $(shell spruce json galaxy.yml | jq '.namespace' | sed 's?"??g')
_name               = $(shell spruce json galaxy.yml | jq '.name'      | sed 's?"??g')
_install_path       = ~/.ansible/collections
_install_path_local = ./collections

### General ###
check-vars:
ifeq ($(strip $(API_KEY)),)
	@echo "API_KEY variable is unset" && false
endif

env-show: check-vars
	@echo "API_KEY=[ set but redacted ]"

versions: check-vars
	@ansible --version

### Galaxy ###
_strip_build_dir:
	@# Temporarily mv dirs for the build so that it is not included in collection.
	mv collections ../collections_tmp || true  # Missing is not an error
	mv venv ../venv_tmp
	@# Check that tmp or hidden files will not be in collection.
	@./bin/check_collection_files.sh

_unstrip_build_dir:
	@# Restore dirs
	mv ../collections_tmp ./collections || true  # Missing is not an error
	mv ../venv_tmp ./venv

galaxy-collection-build: _strip_build_dir
	@# Build the collection.
	rm -rf collections/
	ansible-galaxy collection build
	@make _unstrip_build_dir

galaxy-collection-build-force: _strip_build_dir
	@# Force build the collection. Overwrite an existing collection file.
	ansible-galaxy collection build --force
	@make _unstrip_build_dir

galaxy-collection-publish: check-vars
	@# Publish the collection to https://galaxy.ansible.com/ using the API key provided.
	ansible-galaxy collection publish --api-key $(API_KEY) ./$(_namespace)-$(_name)-$(_version).tar.gz -vvv

galaxy-collection-install:
	@# Download and install from galaxy.ansible.com.
	ansible-galaxy collection install $(_namespace).$(_name) --collections-path $(_install_path) --force

galaxy-collection-install-locally:
	@# Download and install from local tar file.
	ansible-galaxy collection install $(_namespace)-$(_name)-$(_version).tar.gz --collections-path $(_install_path_local)

### Playbooks ###
test-create-resources:
	@# Run test_create_resources.yml as described in DEV_README.md
	cd playbooks && \
		../venv/bin/ansible-playbook --extra-vars "@../ibox_vars/iboxCICD.yaml" \
		--ask-vault-pass test_create_resources.yml 

test-remove-resources:
	@# Run test_remove_resources.yml as described in DEV_README.md
	cd playbooks && \
		../venv/bin/ansible-playbook --extra-vars "@../ibox_vars/iboxCICD.yaml" \
		--ask-vault-pass test_remove_resources.yml 
