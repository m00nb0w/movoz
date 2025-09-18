# Movoz Infrastructure

Infrastructure as Code for the Movoz project using Terraform and AWS.

## 🏗️ Architecture

This project deploys a PostgreSQL database on AWS RDS with the following components:

- **VPC**: Isolated network environment
- **Public Subnets**: For future application load balancers
- **Private Subnets**: For RDS database (no direct internet access)
- **RDS PostgreSQL**: Managed database service
- **Security Groups**: Network access control
- **Backup & Monitoring**: Automated backups and monitoring

## 📁 Project Structure

```
movoz-infra/
├── environments/
│   ├── dev/                    # Development environment config
│   │   └── terraform.tfvars
│   └── prod/                   # Production environment config
│       └── terraform.tfvars
├── modules/                    # Reusable Terraform modules
│   ├── database/
│   └── networking/
├── scripts/
│   └── deploy.sh              # Deployment automation script
├── main.tf                    # Main Terraform configuration
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── versions.tf                # Provider versions
└── README.md                  # This file
```

## 🚀 Quick Start

### Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** installed (>= 1.0)
3. **AWS Account** with RDS permissions

### Installation

```bash
# Install Terraform (macOS)
brew install terraform

# Install AWS CLI (macOS)
brew install awscli

# Configure AWS CLI
aws configure
```

### Deployment

1. **Set the database password**:
```bash
export TF_VAR_db_password="your-secure-password-here"
```

2. **Deploy to development**:
```bash
# Plan the deployment
./scripts/deploy.sh dev plan

# Apply the deployment
./scripts/deploy.sh dev apply
```

3. **Deploy to production**:
```bash
# Plan the deployment
./scripts/deploy.sh prod plan

# Apply the deployment
./scripts/deploy.sh prod apply
```

## 🔧 Manual Terraform Commands

If you prefer to run Terraform commands manually:

```bash
# Initialize Terraform
terraform init

# Select workspace
terraform workspace select dev  # or prod

# Plan deployment
terraform plan -var-file="environments/dev/terraform.tfvars"

# Apply deployment
terraform apply -var-file="environments/dev/terraform.tfvars"

# View outputs
terraform output
```

## 📊 Environment Differences

| Feature | Development | Production |
|---------|-------------|------------|
| Instance Class | db.t3.micro | db.t3.small |
| Storage | 20 GB | 100 GB |
| Backup Retention | 3 days | 30 days |
| Enhanced Monitoring | Disabled | Enabled |
| Performance Insights | Disabled | Enabled |
| Deletion Protection | Disabled | Enabled |
| VPC CIDR | 10.1.0.0/16 | 10.0.0.0/16 |

## 🔐 Security Features

- **Private Subnets**: Database is not publicly accessible
- **Security Groups**: Restrictive network access rules
- **Encryption**: Storage encryption enabled
- **Backups**: Automated daily backups
- **Monitoring**: CloudWatch monitoring (production)

## 📝 Configuration

### Environment Variables

- `TF_VAR_db_password`: Database master password (required)
- `TF_VAR_aws_region`: AWS region (default: us-west-2)

### Customization

Edit the `terraform.tfvars` files in the `environments/` directory to customize:

- Database instance class
- Storage size
- Backup retention
- Maintenance windows

## 🗑️ Cleanup

To destroy the infrastructure:

```bash
# Destroy development environment
./scripts/deploy.sh dev destroy

# Destroy production environment
./scripts/deploy.sh prod destroy
```

## 📋 Outputs

After deployment, you'll get the following outputs:

- `db_endpoint`: Database connection endpoint
- `db_port`: Database port (5432)
- `db_name`: Database name
- `db_username`: Master username
- `vpc_id`: VPC ID
- `subnet_ids`: Private subnet IDs
- `security_group_id`: Database security group ID

## 🔗 Connecting to the Database

```bash
# Get connection details
terraform output

# Connect using psql (if installed)
psql -h <db_endpoint> -p 5432 -U <db_username> -d <db_name>
```

## 🚨 Important Notes

1. **Password Security**: Never commit passwords to version control
2. **Cost Management**: Remember to destroy resources when not needed
3. **Backup Strategy**: Production backups are retained for 30 days
4. **Network Access**: Database is only accessible from within the VPC

## 🛠️ Troubleshooting

### Common Issues

1. **AWS Credentials**: Ensure AWS CLI is configured correctly
2. **Permissions**: Verify your AWS user has RDS permissions
3. **Region**: Check that the specified region supports RDS
4. **Password**: Ensure TF_VAR_db_password is set

### Getting Help

```bash
# Check Terraform version
terraform version

# Validate configuration
terraform validate

# Format code
terraform fmt

# Check AWS credentials
aws sts get-caller-identity
```

## 📈 Future Enhancements

- [ ] Add Redis cache
- [ ] Add Application Load Balancer
- [ ] Add ECS/Fargate for applications
- [ ] Add CloudFront CDN
- [ ] Add monitoring and alerting
- [ ] Add multi-AZ deployment for production


