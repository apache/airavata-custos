#!/bin/bash
# A script that performs a cURL call to the custos server that has the client management API enabled on it.
# This will issue a POST to the endpoint (as per RFC 7591) and will create a new client on the server
# from the given JSON object.
# This register an admin client in the Custos and the Custos admin will manually accept the client, once accepted
# you MUST use the clientId and clientSecret issued in the response for subsequent tenant creations.

source ./setenv.sh

curl -k -X GET -H "Authorization: Bearer $BEARER_TOKEN" $REGISTRATION_URI > output_get.json
cat output_admin.json