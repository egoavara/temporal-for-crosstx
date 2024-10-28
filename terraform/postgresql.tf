resource "helm_release" "postgresql" {
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "postgresql"
  name       = "postgresql"
  version    = "15.5.32"

  namespace        = "default"
  wait             = true
  create_namespace = false
  values = [
    yamlencode({
      global = {
        postgresql = {
          auth = {
            username = "temporal"
            password = "temporal"
          }
        }
      }
      primary = {
        initdb ={
          scripts = {
            "init.sql" = "create table data (id serial primary key, title varchar, contents jsonb);"
          }
        }
        resources = {
          requests = {
            cpu    = "1000m"
            memory = "1000Mi"
          }
          limits = {
            cpu    = "1000m"
            memory = "1000Mi"
          }
        }
        networkPolicy = {
          enabled = false
        }
      }
    })
  ]
}

