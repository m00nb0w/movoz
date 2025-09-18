# GitHub Actions Workflows

This directory contains GitHub Actions workflows for the Movoz infrastructure project.

## 🔄 Available Workflows

### 1. Infrastructure Cost Analysis (`infrastructure-cost-analysis.yml`)

**Triggers:**
- Pull requests that modify `movoz-infra/` files
- Pushes to main/master branch with infrastructure changes

**Features:**
- Calculates cost impact for both dev and prod environments
- Uses Infracost for accurate AWS cost estimation
- Posts cost analysis as PR comments
- Creates alerts for high-cost changes
- Generates detailed cost breakdowns

**Required Secrets:**
- `INFRACOST_API_KEY`: API key for Infracost service
- `AWS_ACCESS_KEY_ID`: AWS access key for cost analysis
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for cost analysis

### 2. Infrastructure Deployment (`infrastructure-deploy.yml`)

**Triggers:**
- Pushes to main/master branch
- Manual workflow dispatch

**Features:**
- Validates Terraform configuration
- Runs security scans with Trivy
- Deploys infrastructure to dev and prod environments
- Supports plan, apply, and destroy actions
- Generates deployment summaries

**Required Secrets:**
- `AWS_ACCESS_KEY_ID`: AWS access key for deployment
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for deployment
- `DB_PASSWORD`: Database password for Terraform

### 3. Infrastructure Cleanup (`infrastructure-cleanup.yml`)

**Triggers:**
- Scheduled: Every Sunday at 2 AM UTC (dev environment only)
- Manual workflow dispatch

**Features:**
- Automatically cleans up development environment weekly
- Manual cleanup for production (with confirmation)
- Cost optimization by removing unused resources
- Generates cleanup summaries

**Required Secrets:**
- `AWS_ACCESS_KEY_ID`: AWS access key for cleanup
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for cleanup
- `DB_PASSWORD`: Database password for Terraform

## 🚀 Getting Started

### 1. Set up Required Secrets

Go to your repository settings → Secrets and variables → Actions, and add:

```bash
# AWS Credentials
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key

# Database Password
DB_PASSWORD=your-secure-database-password

# Infracost API Key (for cost analysis)
INFRACOST_API_KEY=your-infracost-api-key
```

### 2. Get Infracost API Key

1. Sign up at [infracost.io](https://infracost.io)
2. Get your API key from the dashboard
3. Add it to your repository secrets

### 3. Test the Workflows

Create a pull request that modifies files in `movoz-infra/` to trigger the cost analysis workflow.

## 📊 Cost Analysis Features

### Automatic Cost Calculation
- **Real-time pricing**: Uses current AWS pricing data
- **Multi-environment**: Analyzes both dev and prod costs
- **Resource breakdown**: Shows cost per AWS service
- **Change impact**: Calculates cost difference from current state

### Cost Alerts
- **Threshold monitoring**: Alerts when costs exceed $100/month
- **Automatic issues**: Creates GitHub issues for high-cost changes
- **PR comments**: Posts detailed cost analysis on pull requests

### Cost Optimization
- **Weekly cleanup**: Automatically removes dev resources
- **Recommendations**: Suggests cost optimization strategies
- **Resource sizing**: Helps identify over-provisioned resources

## 🛡️ Security Features

### Terraform Validation
- **Format checking**: Ensures consistent code formatting
- **Syntax validation**: Validates Terraform configuration
- **Security scanning**: Uses Trivy for vulnerability detection

### Access Control
- **Environment separation**: Dev and prod use separate workspaces
- **Secret management**: All sensitive data stored in GitHub secrets
- **Audit trail**: All actions logged in GitHub Actions

## 📈 Monitoring and Alerts

### Cost Monitoring
- **Daily cost tracking**: Monitor infrastructure costs
- **Budget alerts**: Get notified of cost increases
- **Resource optimization**: Identify cost-saving opportunities

### Deployment Monitoring
- **Deployment status**: Track successful/failed deployments
- **Resource health**: Monitor AWS resource status
- **Performance metrics**: Track infrastructure performance

## 🔧 Customization

### Environment Configuration
Modify `movoz-infra/environments/*/terraform.tfvars` to customize:
- Instance sizes
- Storage amounts
- Backup retention
- Monitoring settings

### Workflow Triggers
Edit workflow files to customize:
- Trigger conditions
- Environment matrix
- Notification settings
- Cost thresholds

### Cost Thresholds
Update cost alert thresholds in `infrastructure-cost-analysis.yml`:
```yaml
HIGH_COST_THRESHOLD=100  # Alert if monthly cost > $100
```

## 🆘 Troubleshooting

### Common Issues

1. **Missing Secrets**: Ensure all required secrets are set
2. **AWS Permissions**: Verify AWS credentials have necessary permissions
3. **Terraform State**: Check for state file conflicts
4. **Cost Analysis**: Verify Infracost API key is valid

### Debug Steps

1. Check workflow logs in GitHub Actions
2. Verify AWS credentials: `aws sts get-caller-identity`
3. Test Terraform locally: `terraform plan`
4. Validate Infracost: `infracost auth login`

### Getting Help

- Check GitHub Actions logs for detailed error messages
- Review Terraform documentation for configuration issues
- Consult Infracost documentation for cost analysis problems
- Open an issue in the repository for workflow-specific problems

## 📚 Additional Resources

- [Terraform Documentation](https://terraform.io/docs)
- [Infracost Documentation](https://docs.infracost.io)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [AWS RDS Documentation](https://docs.aws.amazon.com/rds)


