provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "k3d-temporal-for-crosstx"
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = "k3d-temporal-for-crosstx"
  }
}

terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }

  backend "kubernetes" {
    config_path    = "~/.kube/config"
    config_context = "k3d-temporal-for-crosstx"
    secret_suffix  = "temporal-for-crosstx"
    namespace      = "default"
  }
}
