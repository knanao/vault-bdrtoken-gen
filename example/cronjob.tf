resource "kubernetes_cron_job_v1" "vault-bdrtoken-gen" {
  metadata {
    name      = "vault-bdrtoken-gen"
    namespace = "vault"
  }
  spec {
    schedule = "0 1-23/7 * * *"
    job_template {
      spec {
        template {
          spec {
            service_account_name = kubernetes_service_account_v1.vault-bdrtoken-gen.metadata.0.name
            container {
              name  = "vault-bdrtoken-gen"
              image = "knanao/vault-bdrtoken-gen:v0.1.0"
              args  = ["--kubernetes-auth-role=vault-bdrtoken-gen", "--batch-token-role=failover-handler", "--bucket-url=gs://vault-bdrtoken"]
              env {
                name  = VAULT_ADDR
                value = "$VAULT_ADDR"
              }
              env {
                name  = TZ
                value = "Asia/Tokyo"
              }
              env {
                name  = VAULT_CACERT
                value = "/vault/userconfig/vault-tls/ca.crt"
              }
              env {
                name  = VAULT_CLIENT_CERT
                value = "/vault/userconfig/vault-tls/tls.crt"
              }
              env {
                name  = VAULT_CLIENT_KEY
                value = "/vault/userconfig/vault-tls/tls.key"
              }
              volume_mount {
                name       = "vault-tls"
                mount_path = "/vault/userconfig/vault-tls"
                read_only  = true
              }
            }
            volume {
              name = "vault-tls"
              secret = {
                default_mode = "0644"
                secret_name  = VAULT_TLS_SECRET_NAME
              }
            }
            restart_policy = "OnFailure"
          }
        }
      }
    }
  }
}

resource "kubernetes_service_account_v1" "vault-bdrtoken-gen" {
  metadata {
    annotations = {
      "iam.gke.io/gcp-service-account" = "vault-bdrtoken-gen@${PROJECT_ID}.iam.gserviceaccount.com"
    }

    name      = "vault-bdrtoken-gen"
    namespace = "vault"
  }
}
