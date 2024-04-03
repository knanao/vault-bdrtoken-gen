package main

import (
	"context"
	"flag"
	"os"
	"time"
	_ "time/tzdata"

	hclog "github.com/hashicorp/go-hclog"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

func main() {
	var kubernetesAuthRole string
	var serviceAccountTokenPath string
	var batchTokenRole string
	var batchTokenTTL string
	var bucketURL string

	flag.StringVar(&kubernetesAuthRole, "kubernetes-auth-role", "", "The role name of the Kubernetes auth.")
	flag.StringVar(&serviceAccountTokenPath, "service-account-token-path", "/var/run/secrets/kubernetes.io/serviceaccount/token", "The path of application's Kubernetes service account token.")
	flag.StringVar(&batchTokenRole, "batch-token-role", "", "The role name of batch DR operation token. The policy for the failover should have been attached to this.")
	flag.StringVar(&batchTokenTTL, "batch-token-ttl", "8h", "Time to live (TTL) for batch DR operation token.")
	flag.StringVar(&bucketURL, "bucket-url", "", "The URL of Cloud Storage. Currently, Google Cloud Storage(GCS) and Amazon S3 are only supported.")
	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "vault-bdrtoken-gen",
		Level: hclog.LevelFromString("INFO"),
	})

	logger.Info("Starting generator")

	now := time.Now()
	var client *vault.TokenAuth
	{
		config := vault.DefaultConfig()
		c, err := vault.NewClient(config)
		if err != nil {
			logger.Error("Unable to initialize Vault client", err)
			os.Exit(1)
		}

		k8sAuth, err := auth.NewKubernetesAuth(
			kubernetesAuthRole,
			auth.WithServiceAccountTokenPath(serviceAccountTokenPath),
		)
		if err != nil {
			logger.Error("Unable to initialize Kubernetes auth method", err)
			os.Exit(1)
		}

		authInfo, err := c.Auth().Login(context.Background(), k8sAuth)
		if err != nil {
			logger.Error("Unable to log in with Kubernetes auth", err)
			os.Exit(1)
		}
		if authInfo == nil {
			logger.Error("No auth info was returned after login")
			os.Exit(1)
		}
		client = c.Auth().Token()
	}

	ctx := context.Background()
	req := &vault.TokenCreateRequest{TTL: batchTokenTTL}
	resp, err := client.CreateWithRoleWithContext(ctx, req, batchTokenRole)
	if err != nil {
		logger.Error("Unable to create token", err)
		os.Exit(1)
	}
	if resp.Auth == nil {
		logger.Error("Unable to retrieve the created token")
		os.Exit(1)
	}

	secret, err := client.Lookup(resp.Auth.ClientToken)
	if err != nil {
		logger.Error("Unable to lookup the token although completed the token created", err)
		os.Exit(1)
	}
	if secret.Data == nil {
		logger.Error("No token data was returned")
		os.Exit(1)
	}

	var (
		issueTime  string
		expireTime string
	)
	if v, ok := secret.Data["issue_time"]; ok {
		if t, err := time.Parse(time.RFC3339, v.(string)); err == nil {
			issueTime = t.Local().Format(time.RFC3339)
		}
	}
	if v, ok := secret.Data["expire_time"]; ok {
		if t, err := time.Parse(time.RFC3339, v.(string)); err == nil {
			expireTime = t.Local().Format(time.RFC3339)
		}
	}
	meta := map[string]string{
		"issue_time":  issueTime,
		"expire_time": expireTime,
	}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		logger.Error("Unable to open the bucket", err)
		os.Exit(1)
	}
	defer bucket.Close()

	file := now.Format(time.RFC3339)
	opts := &blob.WriterOptions{Metadata: meta}
	if err := bucket.WriteAll(ctx, file, []byte(resp.Auth.ClientToken), opts); err != nil {
		logger.Error("Unable to create a file in the bucket", err)
		os.Exit(1)
	}
	logger.Info("Completed batch DR operation token generation")
}
