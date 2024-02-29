# Batch DR Operation Token Generator for Vault Enterprise on Kubernetes
This offers an option to reduce extra labor for Vault Enterprise operator. Vault Enterprise provides Disaster Recovery (DR) replication feature, and this feature is one of the big advantages of using it. If a failure actually occures, currently we have two options to promote a DR secondary cluster to a new primary. The first is to use a DR operation token, and the other is a batch DR operation token. Don't mention the differences here, so please read [this docs](https://developer.hashicorp.com/vault/tutorials/enterprise/disaster-recovery#dr-operation-token-strategy). For most cases, a batch DR operation token should be preferable because it has explicit TTL, and no necessary to scramble to prepare for unseal keys or recovery keys at that time. Especially, it will shine when these operations are outsourced as a DR operation token needs an extra step to revoke the token after the operation.

# Usage
This supports only for Kubernetes Auth so far, and Google Cloud Storage(GCS) and Amazon S3 are only supported as the token storage. Please see [examples](./example/cronjob.yaml) for more details.

## Congigurations
- `--kubernetes-auth-role` `(string: required)` - Role to use for authenticating to Vault.
- `--service-account-token-path` `(string: "/var/run/secrets/kubernetes.io/serviceaccount/token")` - Path to where your application's Kubernetes service account token is mounted,
- `--batch-token-role` `(string: required)` - Token role for failover operations. Its type should be `batch`.
- `--batch-token-ttl` `(string: "8h")` - TTL of the batch DR operation token. When executing as Cronjob, make sure that the overlap period of two or more tokens is long enough for the DR operation.
- `--bucket-url` `(string: required)`- The URL of Cloud Storage. Currently, Google Cloud Storage(GCS) and Amazon S3 are only supported. e.g. `gs://vault-bdrtoken`

## Environment variables

### `VAULT_ADDR`
Address of the Vault server expressed as a URL and port, for example:
`https://127.0.0.1:8200/`.

### `TZ`
Timezone for the file name and its metadata. Please see the supported [zones](https://github.com/arp242/tz/blob/bf333631bec4/list.go), for example:
`Asia/Tokyo`.
