#!/bin/bash

# Movoz Infrastructure Deployment Script
# Usage: ./deploy.sh <environment> <action>
# Example: ./deploy.sh dev plan
# Example: ./deploy.sh prod apply

set -e

ENVIRONMENT=$1
ACTION=$2

if [ -z "$ENVIRONMENT" ] || [ -z "$ACTION" ]; then
    echo "Usage: $0 <environment> <action>"
    echo "Environments: dev, prod"
    echo "Actions: plan, apply, destroy"
    exit 1
fi

if [ "$ENVIRONMENT" != "dev" ] && [ "$ENVIRONMENT" != "prod" ]; then
    echo "Error: Environment must be 'dev' or 'prod'"
    exit 1
fi

if [ "$ACTION" != "plan" ] && [ "$ACTION" != "apply" ] && [ "$ACTION" != "destroy" ]; then
    echo "Error: Action must be 'plan', 'apply', or 'destroy'"
    exit 1
fi

# Check if AWS CLI is configured
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "Error: AWS CLI not configured. Please run 'aws configure' first."
    exit 1
fi

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "Error: Terraform not installed. Please install Terraform first."
    exit 1
fi

# Set environment variables
export TF_VAR_environment=$ENVIRONMENT
export TF_VAR_aws_region=$(aws configure get region)

# Check if db_password is set
if [ -z "$TF_VAR_db_password" ]; then
    echo "Error: TF_VAR_db_password environment variable not set."
    echo "Please set it with: export TF_VAR_db_password='your-secure-password'"
    exit 1
fi

# Navigate to the project root
cd "$(dirname "$0")/.."

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Select workspace
echo "Selecting workspace: $ENVIRONMENT"
terraform workspace select $ENVIRONMENT 2>/dev/null || terraform workspace new $ENVIRONMENT

# Load environment-specific variables
if [ -f "environments/$ENVIRONMENT/terraform.tfvars" ]; then
    echo "Loading environment variables from environments/$ENVIRONMENT/terraform.tfvars"
    terraform plan -var-file="environments/$ENVIRONMENT/terraform.tfvars" -out="$ENVIRONMENT.tfplan"
else
    echo "Warning: No terraform.tfvars file found for environment $ENVIRONMENT"
fi

# Execute the action
case $ACTION in
    "plan")
        echo "Planning infrastructure for $ENVIRONMENT environment..."
        terraform plan -var-file="environments/$ENVIRONMENT/terraform.tfvars"
        ;;
    "apply")
        echo "Applying infrastructure for $ENVIRONMENT environment..."
        terraform apply -var-file="environments/$ENVIRONMENT/terraform.tfvars" -auto-approve
        echo "Deployment completed successfully!"
        echo "Database endpoint: $(terraform output -raw db_endpoint)"
        ;;
    "destroy")
        echo "WARNING: This will destroy all infrastructure in $ENVIRONMENT environment!"
        read -p "Are you sure? Type 'yes' to continue: " confirm
        if [ "$confirm" = "yes" ]; then
            terraform destroy -var-file="environments/$ENVIRONMENT/terraform.tfvars" -auto-approve
            echo "Infrastructure destroyed successfully!"
        else
            echo "Destroy cancelled."
        fi
        ;;
esac
