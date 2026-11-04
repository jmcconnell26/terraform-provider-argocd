package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetCertificateOptsWritesInlineValues(t *testing.T) {
	t.Parallel()

	p := ArgoCDProviderConfig{
		Cert:          types.StringValue("ca-pem"),
		ClientCert:    types.StringValue("cert-pem"),
		ClientCertKey: types.StringValue("key-pem"),
	}
	opts := &apiclient.ClientOptions{}

	if diags := p.setCertificateOpts(opts); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	t.Cleanup(func() { os.RemoveAll(filepath.Dir(opts.CertFile)) })

	for path, want := range map[string]string{
		opts.CertFile:          "ca-pem",
		opts.ClientCertFile:    "cert-pem",
		opts.ClientCertKeyFile: "key-pem",
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %q: %v", path, err)
		}

		if string(got) != want {
			t.Errorf("%q contains %q, want %q", path, got, want)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %q: %v", path, err)
		}

		if info.Mode().Perm() != 0o600 {
			t.Errorf("%q has mode %v, want 0600", path, info.Mode().Perm())
		}
	}

	if filepath.Dir(opts.CertFile) != filepath.Dir(opts.ClientCertKeyFile) {
		t.Errorf("expected all certificates in one directory, got %q and %q", opts.CertFile, opts.ClientCertKeyFile)
	}
}

func TestSetCertificateOptsPrefersFilePaths(t *testing.T) {
	t.Parallel()

	p := ArgoCDProviderConfig{
		CertFile:          types.StringValue("/etc/ca.crt"),
		Cert:              types.StringValue("ignored"),
		ClientCertFile:    types.StringValue("/etc/client.crt"),
		ClientCert:        types.StringValue("ignored"),
		ClientCertKeyFile: types.StringValue("/etc/client.key"),
		ClientCertKey:     types.StringValue("ignored"),
	}
	opts := &apiclient.ClientOptions{}

	if diags := p.setCertificateOpts(opts); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if opts.CertFile != "/etc/ca.crt" || opts.ClientCertFile != "/etc/client.crt" || opts.ClientCertKeyFile != "/etc/client.key" {
		t.Errorf("file paths not passed through: %+v", opts)
	}
}

func TestSetCertificateOptsLeavesUnsetEmpty(t *testing.T) {
	t.Parallel()

	opts := &apiclient.ClientOptions{}

	if diags := (ArgoCDProviderConfig{}).setCertificateOpts(opts); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if opts.CertFile != "" || opts.ClientCertFile != "" || opts.ClientCertKeyFile != "" {
		t.Errorf("expected no certificate paths, got %+v", opts)
	}
}
