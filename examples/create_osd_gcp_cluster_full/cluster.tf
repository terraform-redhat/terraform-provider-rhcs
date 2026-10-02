###############################################################################
# OSD cluster: WIF + CMEK + Shared VPC + PSC + private API + inline admin user
###############################################################################
# Trial product (osdtrial) — 60-day evaluation, no OSD service fee.
# GCP infra costs (VMs, disks, NAT, KMS) still apply.
#
# The htpasswd cluster-admin user is baked into the cluster create body via
# the admin_credentials field (mirroring rhcs_cluster_rosa_classic). No
# separate IDP resource needed.

resource "rhcs_cluster_osd_gcp" "osd" {
  depends_on = [
    module.wif_gcp,
    google_kms_crypto_key_iam_member.kms_sa,
    google_kms_crypto_key_iam_member.compute_agent,
  ]

  name           = var.cluster_name
  product        = var.product
  billing_model  = var.billing_model
  cloud_region   = var.region
  gcp_project_id = var.service_project_id
  wif_config_id  = rhcs_wif_config.wif.id
  version        = var.openshift_version
  compute_nodes  = var.compute_nodes
  ccs_enabled    = true
  private        = true

  gcp_network = {
    vpc_name             = var.network_name
    vpc_project_id       = var.host_project_id
    compute_subnet       = var.compute_subnet
    control_plane_subnet = var.control_plane_subnet
  }

  private_service_connect = {
    service_attachment_subnet = var.psc_subnet
  }

  gcp_encryption_key = {
    kms_key_service_account = google_service_account.kms.email
    key_location            = var.region
    key_name                = google_kms_crypto_key.osd.name
    key_ring                = google_kms_key_ring.osd.name
  }

  security = {
    secure_boot = true
  }

  admin_credentials = {
    username = var.admin_username
    password = var.admin_password != "" ? var.admin_password : null
  }
}
