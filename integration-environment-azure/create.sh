#!/bin/bash
set -e
cd $(dirname "$0")

# Setup Auth for Azure RM provider in terraform
export ARM_CLIENT_ID=$DATABRICKS_AZURE_CLIENT_ID
export ARM_CLIENT_SECRET=$DATABRICKS_AZURE_CLIENT_SECRET
export ARM_SUBSCRIPTION_ID=$DATABRICKS_AZURE_SUBSCRIPTION_ID
export ARM_TENANT_ID=$DATABRICKS_AZURE_TENANT_ID

terraform apply -auto-approve

# Save the details of the resource to json files for use by the tests
terraform output -json workspace > workspace.json
terraform output -json datalake > datalake.json