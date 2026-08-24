resource "rhcs_default_ingress" "default_ingress" {
  cluster = "cluster-id-123"
  component_routes = {
    console = {
      hostname       = "console.example.com"
      tls_secret_ref = "console-tls-secret"
    }
  }
}
