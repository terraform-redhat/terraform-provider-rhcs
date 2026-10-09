###############################################################################
# CMEK: KMS keyring + key + dedicated SA + two IAM grants
###############################################################################
# OCM independently validates two principals have encrypter/decrypter on the key:
#   - kms_key_service_account passed to rhcs_cluster_osd_gcp (we create it)
#   - the Compute Engine Service Agent (created lazily by GCP; exists once
#     Compute API is enabled, which the WIF module triggers above)

resource "google_kms_key_ring" "osd" {
  project  = var.service_project_id
  name     = "${var.cluster_name}-keyring"
  location = var.region
}

resource "google_kms_crypto_key" "osd" {
  name     = "${var.cluster_name}-key"
  key_ring = google_kms_key_ring.osd.id
  purpose  = "ENCRYPT_DECRYPT"
}

resource "google_service_account" "kms" {
  account_id   = "${var.cluster_name}-kms"
  display_name = "CMEK access for OSD cluster ${var.cluster_name}"
  project      = var.service_project_id
}

resource "google_kms_crypto_key_iam_member" "kms_sa" {
  crypto_key_id = google_kms_crypto_key.osd.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${google_service_account.kms.email}"
}

resource "google_kms_crypto_key_iam_member" "compute_agent" {
  crypto_key_id = google_kms_crypto_key.osd.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.service.number}@compute-system.iam.gserviceaccount.com"
}
