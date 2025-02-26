terraform {
  required_version = "1"
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "6.22.0"
    }
    kubernetes = ">=2.30.0"
  }
}

