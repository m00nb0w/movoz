# Configure the AWS Provider
provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "movoz"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

# VPC
resource "aws_vpc" "movoz_vpc" {
  cidr_block           = var.environment == "prod" ? "10.0.0.0/16" : "10.1.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "movoz-${var.environment}-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "movoz_igw" {
  vpc_id = aws_vpc.movoz_vpc.id

  tags = {
    Name = "movoz-${var.environment}-igw"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count = 2

  vpc_id                  = aws_vpc.movoz_vpc.id
  cidr_block              = var.environment == "prod" ? "10.0.${count.index + 1}.0/24" : "10.1.${count.index + 1}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "movoz-${var.environment}-public-subnet-${count.index + 1}"
    Type = "public"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count = 2

  vpc_id            = aws_vpc.movoz_vpc.id
  cidr_block        = var.environment == "prod" ? "10.0.${count.index + 10}.0/24" : "10.1.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "movoz-${var.environment}-private-subnet-${count.index + 1}"
    Type = "private"
  }
}

# Route Table for Public Subnets
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.movoz_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.movoz_igw.id
  }

  tags = {
    Name = "movoz-${var.environment}-public-rt"
  }
}

# Route Table Associations for Public Subnets
resource "aws_route_table_association" "public" {
  count = 2

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# DB Subnet Group
resource "aws_db_subnet_group" "movoz_db_subnet_group" {
  name       = "movoz-${var.environment}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "movoz-${var.environment}-db-subnet-group"
  }
}

# Security Group for RDS
resource "aws_security_group" "db_sg" {
  name_prefix = "movoz-${var.environment}-db-sg"
  vpc_id      = aws_vpc.movoz_vpc.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.movoz_vpc.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "movoz-${var.environment}-db-sg"
  }
}

# RDS Instance
resource "aws_db_instance" "movoz_db" {
  identifier = "movoz-${var.environment}-db"

  # Engine
  engine         = var.db_engine
  engine_version = var.db_engine_version

  # Instance
  instance_class    = var.db_instance_class
  allocated_storage = var.db_allocated_storage
  storage_type      = "gp2"

  # Database
  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  # Network
  db_subnet_group_name   = aws_db_subnet_group.movoz_db_subnet_group.name
  vpc_security_group_ids = [aws_security_group.db_sg.id]
  publicly_accessible    = false

  # Backup
  backup_retention_period = var.backup_retention_period
  backup_window          = var.backup_window
  maintenance_window     = var.maintenance_window

  # Monitoring
  monitoring_interval = var.environment == "prod" ? 60 : 0
  monitoring_role_arn = var.environment == "prod" ? aws_iam_role.rds_enhanced_monitoring[0].arn : null

  # Security
  storage_encrypted = true
  deletion_protection = var.environment == "prod" ? true : false

  # Performance Insights
  performance_insights_enabled = var.environment == "prod" ? true : false

  tags = {
    Name = "movoz-${var.environment}-db"
  }
}

# IAM Role for Enhanced Monitoring (Production only)
resource "aws_iam_role" "rds_enhanced_monitoring" {
  count = var.environment == "prod" ? 1 : 0

  name = "movoz-${var.environment}-rds-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "movoz-${var.environment}-rds-monitoring-role"
  }
}

resource "aws_iam_role_policy_attachment" "rds_enhanced_monitoring" {
  count = var.environment == "prod" ? 1 : 0

  role       = aws_iam_role.rds_enhanced_monitoring[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}


