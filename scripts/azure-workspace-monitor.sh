#!/bin/bash

set -e

# Onetime setup (or use devcontainer which has these installed - see readme for devcontainer instructions)
# 1. Install Az CLI
# 2. Add databricks addin `az extension add --name databricks`

RG=inttestworkspace4
LOCATION=eastus

az group create --name $RG --location $LOCATION

for i in {1..5}
do
    # Create 10 workspaces in parrallel
    az databricks workspace create --resource-group $RG --name inttest$RANDOM --location $LOCATION --sku standard --no-wait &
done

## todo get PAT
tenantId=$(az account show --query tenantId -o tsv)

# Get a token for the global Databricks application.
# The resource name is fixed and never changes.
token_response=$(az account get-access-token --resource 2ff814a6-3304-4ab8-85cb-cd0e6f879c1d)
token=$(jq .accessToken -r <<< "$token_response")

# Get a token for the Azure management API
token_response=$(az account get-access-token --resource https://management.core.windows.net/)
azToken=$(jq .accessToken -r <<< "$token_response")

## While true
while true
do
    for wsid in `az databricks workspace list -g $RG --query "[].id" -o tsv`
    do 
        LOGNAME=`echo $wsid | base64`.log
        echo -e "\n`date` \n" | tee -a $LOGNAME

        echo "Fetching clusters/list-node-types" | tee -a $LOGNAME
        curl https://$LOCATION.azuredatabricks.net/api/2.0/clusters/list-node-types \
            -H "Authorization: Bearer $token" \
            -H "X-Databricks-Azure-SP-Management-Token:$azToken" \
            -H "X-Databricks-Azure-Workspace-Resource-Id:$wsid" | tee -a $LOGNAME
        
        echo -e "\n`date` \n" | tee -a $LOGNAME

        echo "Fetching clusters/list" | tee -a $LOGNAME
        curl https://$LOCATION.azuredatabricks.net/api/2.0/clusters/list \
            -H "Authorization: Bearer $token" \
            -H "X-Databricks-Azure-SP-Management-Token:$azToken" \
            -H "X-Databricks-Azure-Workspace-Resource-Id:$wsid" | tee -a $LOGNAME
    done
    sleep 1
done
