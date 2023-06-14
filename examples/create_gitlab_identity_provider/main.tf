#
# Copyright (c) 2023 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.67"
    }
    ocm = {
      version = ">=0.0.5"
      source  = "terraform.local/local/ocm"
    }
  }
}

provider "ocm" {
token = "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJhZDUyMjdhMy1iY2ZkLTRjZjAtYTdiNi0zOTk4MzVhMDg1NjYifQ.eyJpYXQiOjE2ODUzNjIwODEsImp0aSI6IjA3NjVkMmE4LTIwYjktNGEwOC04ZmQ3LWI4NzI3MTIwMTIxMiIsImlzcyI6Imh0dHBzOi8vc3NvLnJlZGhhdC5jb20vYXV0aC9yZWFsbXMvcmVkaGF0LWV4dGVybmFsIiwiYXVkIjoiaHR0cHM6Ly9zc28ucmVkaGF0LmNvbS9hdXRoL3JlYWxtcy9yZWRoYXQtZXh0ZXJuYWwiLCJzdWIiOiJmOjUyOGQ3NmZmLWY3MDgtNDNlZC04Y2Q1LWZlMTZmNGZlMGNlNjpoYWMtZWNvc3lzdGVtIiwidHlwIjoiT2ZmbGluZSIsImF6cCI6ImNsb3VkLXNlcnZpY2VzIiwibm9uY2UiOiJhOWFlNzczNC1iNDA0LTQwMTAtYjY2Mi0xNWY4NjZhOTU2M2UiLCJzZXNzaW9uX3N0YXRlIjoiOTM2MWJkYjQtZGZiYS00NTk2LWJlM2YtMWU5YTYwMjc2NDA1Iiwic2NvcGUiOiJvcGVuaWQgYXBpLmlhbS5zZXJ2aWNlX2FjY291bnRzIGFwaS5pYW0ub3JnYW5pemF0aW9uIG9mZmxpbmVfYWNjZXNzIiwic2lkIjoiOTM2MWJkYjQtZGZiYS00NTk2LWJlM2YtMWU5YTYwMjc2NDA1In0.NECrUAFktGhmPBo7zyWGtot8GebpTRbqDnSuujhSB3I"
 url="https://api.stage.openshift.com"
}

resource "ocm_identity_provider" "google_idp" {
  cluster = "249kjmgpkdnmbe6cgnm6gjeloshs0hd6"
  name    = "testclient1"
  google = {
    client_id     = "576250202330-lmk5m0eog3c1o0da8dg09g15qjccbtsk.apps.googleusercontent.com"
    client_secret = "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"
    hosted_domain = "redhat.com"
  }
}