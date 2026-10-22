terraform {
  required_version = ">= 1.5"
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "~> 0.6"
    }
  }
}

# The provider reads the API key from the NEON_API_KEY environment variable.
provider "neon" {}

variable "org_id" {
  type    = string
  default = "org-twilight-fire-15727881"
}

variable "region_id" {
  type    = string
  default = "aws-us-east-2"
}

resource "neon_project" "user_auth" {
  name                      = "user-auth"
  org_id                    = var.org_id
  region_id                 = var.region_id
  pg_version                = 17
  history_retention_seconds = 21600 # 6 hours, the free-plan maximum

  branch {
    name          = "main"
    database_name = "user_auth"
    role_name     = "user_auth"
  }
}

# Pooled connection string for AUTH_DATABASE_URL. Read it with: terraform output -raw database_url
output "database_url" {
  value     = neon_project.user_auth.connection_uri_pooler
  sensitive = true
}
