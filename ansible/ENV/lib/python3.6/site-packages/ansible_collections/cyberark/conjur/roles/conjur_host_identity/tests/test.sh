#!/bin/bash -e
set -x

function finish {
  echo 'Removing test environment'
  echo '---'
  docker-compose down -v
  rm -rf inventory.tmp
}
trap finish EXIT
finish

# normalises project name by filtering non alphanumeric characters and transforming to lowercase
declare -x COMPOSE_PROJECT_NAME
COMPOSE_PROJECT_NAME=$(echo "${BUILD_TAG:-ansible-plugin-testing}-conjur-host-identity" | sed -e 's/[^[:alnum:]]//g' | tr '[:upper:]' '[:lower:]')

declare -x ANSIBLE_CONJUR_AUTHN_API_KEY=''
declare -x CLI_CONJUR_AUTHN_API_KEY=''
declare cli_cid=''
declare conjur_cid=''
declare ansible_cid=''

function api_key_for {
  local role_id=$1
  if [ ! -z "$role_id" ]
  then
    docker exec ${conjur_cid} rails r "print Credentials['${role_id}'].api_key"
  else
    echo ERROR: api_key_for called with no argument 1>&2
    exit 1
  fi
}

function hf_token {
  docker exec ${cli_cid} conjur hostfactory tokens create \
    --duration-days=5 \
    ansible/ansible-factory | jq -r '.[0].token'
}

function setup_conjur {
  echo "---- setting up conjur ----"
  # run policy
  docker exec ${cli_cid} conjur policy load root /policy/root.yml

  # set secret values
  docker exec ${cli_cid} bash -c '
    conjur variable values add ansible/target-password target_secret_password
  '
}

function run_test_cases {
  for test_case in test_cases/*; do
    teardown_and_setup
    run_test_case "$(basename -- "$test_case")"
  done
}

function run_test_case {
  echo "---- testing ${test_case} ----"
  local test_case=$1
  if [ ! -z "$test_case" ]
  then
    docker exec "${ansible_cid}" env HFTOKEN="$(hf_token)" bash -c "
      cd tests
      ansible-playbook test_cases/${test_case}/playbook.yml
    "
    docker exec "${ansible_cid}" bash -c "
      cd tests
      py.test --junitxml=./junit/${test_case} --connection docker -v test_cases/${test_case}/tests/test_default.py
    "
  else
    echo ERROR: run_test called with no argument 1>&2
    exit 1
  fi
}

function teardown_and_setup {
  docker-compose up -d --force-recreate --scale test_app_ubuntu=2 test_app_ubuntu
  docker-compose up -d --force-recreate --scale test_app_centos=2 test_app_centos
}

function wait_for_server {
  # shellcheck disable=SC2016
  docker exec "${cli_cid}" bash -c '
    for i in $( seq 20 ); do
      curl -o /dev/null -fs -X OPTIONS ${CONJUR_APPLIANCE_URL} > /dev/null && echo "server is up" && break
      echo "."
      sleep 2
    done
  '
}

function fetch_ssl_cert {
  (docker-compose exec -T conjur-proxy-nginx cat cert.crt) > conjur.pem
}

function generate_inventory {
  # uses .j2 template to generate inventory prepended with COMPOSE_PROJECT_NAME
  docker-compose exec -T ansible bash -c '
    cd tests
    ansible-playbook inventory-playbook.yml
  '
}

function main() {
  docker-compose up -d --build
  generate_inventory

  conjur_cid=$(docker-compose ps -q conjur)
  cli_cid=$(docker-compose ps -q conjur_cli)
  fetch_ssl_cert
  wait_for_server

  CLI_CONJUR_AUTHN_API_KEY=$(api_key_for 'cucumber:user:admin')
  docker-compose up -d conjur_cli
  cli_cid=$(docker-compose ps -q conjur_cli)
  setup_conjur

  ANSIBLE_CONJUR_AUTHN_API_KEY=$(api_key_for 'cucumber:host:ansible/ansible-master')
  docker-compose up -d ansible
  ansible_cid=$(docker-compose ps -q ansible)

  run_test_cases
}

main
