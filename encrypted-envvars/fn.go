package example

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"

	kmsapi "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

var username string
var password string

// cryptoKeyID is the GCP crypto key ID to decrypt the data. It is guarded by
// IAM, so it's safe to have in plaintext.
var cryptoKeyID = os.Getenv("KMS_CRYPTO_KEY_ID")

func init() {
	ctx := context.Background()
	kms, err := kmsapi.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	username, err = decrypt(kms, os.Getenv("DB_USER"))
	if err != nil {
		log.Fatalf("failed to decrypt username: %s", err)
	}

	password, err = decrypt(kms, os.Getenv("DB_PASS"))
	if err != nil {
		log.Fatalf("failed to decrypt password: %s", err)
	}
}

func decrypt(kms *kmsapi.KeyManagementClient, s string) (string, error) {
	if s == "" {
		return "", errors.New("data is empty")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode as base64")
	}

	ctx := context.Background()
	resp, err := kms.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       cryptoKeyID,
		Ciphertext: ciphertext,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to decrypt")
	}
	return strings.TrimSpace(string(resp.Plaintext)), nil
}

func Secrets(w http.ResponseWriter, r *http.Request) {
	// You could use a sync.Once and decrypt on the first call (instead of //
	// function initialization), or you could create a new client and decrypt
	// on every call depending on how performant your function needs to be.
	fmt.Fprintf(w, username+":"+password)
}