#!/bin/bash
# This file contains the environment variables for the service. Set them here and they should
# get picked up by each script as needed (this assumes everything is being run from the current
# directory).

# Tip: If you are working with several different clients, you may want to comment out the
# setting REGISTRATION_URI so it does not get set to what is here.


export SERVER=https://custos.scigap.org:32036/tenant-management/v1.0.0/oauth2/tenant
export ACTIVATION_ENDPOINT=https://custos.scigap.org:32036/tenant-management/v1.0.0/status
export ADMIN_ID=custos/In634rvBEjIWtUHUsjF9/10000307
export ADMIN_SECRET=GgDv0GiXWsMZeqqQINMiTUtzLX6zpdIMpmkWpsoc
export REGISTRATION_URI=https://custos.scigap.org:32036/tenant-management/v1.0.0/oauth2/tenant?client_id=custos/h9RWrJZTxFoewlVtZf0x/10000308

# We set the bearer token here so it is available subsequently. This is the least problematic way to
# do this since it is easy to get the escaping wrong.

#export BEARER_TOKEN=$(echo -n $ADMIN_ID:$ADMIN_SECRET | base64 -w 0)
export BEARER_TOKEN=Y3VzdG9zL0luNjM0cnZCRWpJV3RVSFVzakY5LzEwMDAwMzA3OkdnRHYwR2lYV3NNWmVxcVFJTk1pVFV0ekxYNnpwZElNcG1rV3Bzb2M=