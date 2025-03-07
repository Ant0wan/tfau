module "labeler" {
  source = "git@github.com:PrestaShopCorp/terraform-labeler.git?ref=v1.5.0"
}

module "registry" {
  source = "git@github.com:PrestaShopCorp/terraform-google-artifact-registry.git?ref=v1.2.0"
}

module "redis" {
  source                  = "terraform-google-modules/memorystore/google"
  version                 = "~>13.3"
}

module "buckets" {
  source          = "terraform-google-modules/cloud-storage/google"
  version         = "~>9.1"
}

module "bucket_oathkeeper_rules" {
  source           = "terraform-google-modules/cloud-storage/google"
  version          = "~>9.1"
}

module "datastore_backup" {
  source = "git@github.com:PrestaShopCorp/terraform-google-datastore-backup.git?ref=v1.4.0"
}

module "billing_postgresql" {
  source  = "GoogleCloudPlatform/sql-db/google//modules/postgresql"
  version = "~>25.2"
}

