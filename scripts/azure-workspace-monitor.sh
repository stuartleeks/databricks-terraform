#!/bin/bash

set -e

# Onetime setup (or use devcontainer which has these installed - see readme for devcontainer instructions)
# 1. Install Az CLI
# 2. Add databricks addin `az extension add --name databricks`

RG=inttestworkspace10
LOCATION=eastus

az group create --name $RG --location $LOCATION

for i in {1..3}
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
        # Create a unique log file name from the id
        LOGNAME="`echo $wsid | base64`.log"
        # Get workspace provisioning status
        status=`az resource show --id $wsid --query properties.provisioningState -o tsv`
        echo -e "\nIteration:`date` workspaceStatus: $status" | tee -a $LOGNAME

        # Get the workspace url
        workspaceURL=`az resource show --id $wsid --query properties.workspaceUrl -o tsv`
        # Disabled use for comparison using location based url
        #$workspaceURL = "$LOCATION.azuredatabricks.net"

        echo "clusters/list-node-types" | tee -a $LOGNAME
        curl -i https://$workspaceURL/api/2.0/clusters/list-node-types \
            -H "Authorization: Bearer $token" \
            -H "X-Databricks-Azure-SP-Management-Token:$azToken" \
            -H "X-Databricks-Azure-Workspace-Resource-Id:$wsid" | tee -a $LOGNAME
        
        echo "clusters/create" | tee -a $LOGNAME
        curl -i https://$workspaceURL/api/2.0/clusters/create \
            -H "Authorization: Bearer $token" \
            -H "X-Databricks-Azure-SP-Management-Token:$azToken" \
            -H "X-Databricks-Azure-Workspace-Resource-Id:$wsid" \
            -H 'Content-Type: application/json' \
            -X POST \
            -d '{
  "cluster_name": "high-concurrency-cluster",
  "spark_version": "6.4.x-scala2.11",
  "node_type_id": "Standard_DS3_v2",
  "autotermination_minutes":10,
  "start_cluster": false,
  "num_workers": 0
}'| tee -a $LOGNAME

    done
    sleep 1
done
