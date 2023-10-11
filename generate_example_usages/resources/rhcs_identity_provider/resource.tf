# htpasswd
resource "rhcs_identity_provider" "htpasswd_idp" {
  cluster = "<cluster-id>"
  name    = "htpasswd"
  htpasswd = {
    users = [{
      username = "user1"
      password = "423!@edf#@S29e!"
    },
    ]
  }
}

# github
resource "ocm_identity_provider" "github_idp" {
  cluster = "<cluster-id>"
  name    = "Github"
  github = {
    client_id     = "<client-id>"
    client_secret = "<client-secret>"
    organizations = ["<org>"]
  }
}

# gitlab
resource "ocm_identity_provider" "gitlab_idp" {
  cluster = "<cluster-id>"
  name    = "GitLab"
  gitlab = {
    client_id     = "<client-id>"
    client_secret = "<client-secret>"
    url           = "<gitlab-url>"
  }
}

# google
resource "ocm_identity_provider" "google_idp" {
  cluster = "<cluster-id>"
  name = "google"
  google = {
    client_id = "<id>"
    client_secret = "<secret>"
    hosted_domain = "<hosted-domain>"
  }
}

# ldap
resource "rhcs_identity_provider" "ldap_idp" {
  cluster = "<cluster-id>"
  name    = "ldap"
  ldap = {
    url        = "<ldap-url>"
    attributes = {}
    # Optional Attributes
    ca       = "<ldap-ca>"
    insecure = true
  }
}

