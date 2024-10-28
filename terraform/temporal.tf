

resource "helm_release" "temporal" {
  depends_on = [helm_release.postgresql]
  chart      = "https://github.com/temporalio/helm-charts/releases/download/temporal-0.46.2/temporal-0.46.2.tgz"
  name       = "temporal"

  namespace        = "default"
  create_namespace = false
  wait             = true
  values = [ # disable all batteries
    yamlencode({
      cassandra = {
        enabled = false
      },
      mysql = {
        enabled = false
      },
      postgresql = {
        enabled = true
      },
      prometheus = {
        enabled = false
      },
      grafana = {
        enabled = false
      },
      elasticsearch = {
        enabled = false
      },
    }),
    # schema setup
    yamlencode({
      schema = {
        createDatabase = {
          enabled = true
        }
        setup = {
          enabled = true
        }
        update = {
          enabled = true
        }
      }
    }),
    # config
    yamlencode({
      server = {
        replicaCount = 1

        config = {
          persistence = {
            default = {
              driver = "sql"
              sql = {
                driver          = "postgres12"
                host            = "postgresql.default.svc"
                port            = 5432
                database        = "temporal"
                user            = "temporal"
                password        = "temporal"
                maxConns        = 20
                maxConnLifetime = "1h"
              }
            }
            visibility = {
              driver = "sql"
              sql = {
                driver          = "postgres12"
                host            = "postgresql.default.svc"
                port            = 5432
                database        = "temporal_visibility"
                user            = "temporal"
                password        = "temporal"
                maxConns        = 20
                maxConnLifetime = "1h"
              }
            }
          }
        }
      }
    })
  ]
}
