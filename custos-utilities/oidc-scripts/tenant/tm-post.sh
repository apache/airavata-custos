#!/bin/bash
# A script that performs a cURL call to the custos server that has the client management API enabled on it.
# This will issue a POST to the endpoint (as per RFC 7591) and will create a new client on the server
# from the given JSON object.
# This register a  tenant under the admin tenant credentials in the Custos and the Custos tenant will be automatically accepted

source ./setenv.sh

curl -k -X POST -H "Authorization: Bearer $BEARER_TOKEN"  -H "Content-Type: application/json; charset=UTF-8" --data @$1 $SERVER > output_tm_creation.json
cat output_admin.json