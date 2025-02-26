module "my_module" {
  version = "1.2.3"
}

module "another_module" {
  version = "4.5.6"
}

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "6.22.0"
    }
  }
}
