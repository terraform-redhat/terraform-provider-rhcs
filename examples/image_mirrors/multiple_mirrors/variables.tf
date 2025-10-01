variable "cluster_id" {
  type        = string
  description = "The ID of the ROSA HCP cluster where the image mirrors will be created"
}

variable "registry_mirrors" {
  type        = map(list(string))
  description = "Map of source registries to their mirror lists"
  default = {
    "docker.io/library/nginx" = [
      "quay.io/my-org/nginx",
      "registry.example.com/nginx"
    ]
    "docker.io/library/redis" = [
      "quay.io/my-org/redis",
      "registry.example.com/redis",
      "docker.io/backup/redis"
    ]
    "docker.io/library/postgres" = [
      "registry.example.com/postgres",
      "quay.io/backup/postgres"
    ]
    "quay.io/prometheus/prometheus" = [
      "registry.corp.example.com/prometheus/prometheus"
    ]
  }
}