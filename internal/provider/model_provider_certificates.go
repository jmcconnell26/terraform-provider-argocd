package provider

import (
	"os"
	"path/filepath"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (p ArgoCDProviderConfig) setCertificateOpts(opts *apiclient.ClientOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	w := certificateWriter{}

	opts.CertFile, diags = w.resolve(p.CertFile, p.Cert, "ca.crt", diags)
	opts.ClientCertFile, diags = w.resolve(p.ClientCertFile, p.ClientCert, "client.crt", diags)
	opts.ClientCertKeyFile, diags = w.resolve(p.ClientCertKeyFile, p.ClientCertKey, "client.key", diags)

	return diags
}

type certificateWriter struct {
	dir string
}

func (w *certificateWriter) resolve(path, contents types.String, name string, diags diag.Diagnostics) (string, diag.Diagnostics) {
	if v := path.ValueString(); v != "" {
		return v, diags
	}

	v := contents.ValueString()
	if v == "" {
		return "", diags
	}

	if w.dir == "" {
		dir, err := os.MkdirTemp("", "terraform-provider-argocd-")
		if err != nil {
			diags.Append(diagnostics.Error("failed to create temporary certificate directory", err)...)
			return "", diags
		}

		w.dir = dir
	}

	fp := filepath.Join(w.dir, name)
	if err := os.WriteFile(fp, []byte(v), 0o600); err != nil {
		diags.Append(diagnostics.Error("failed to write "+name, err)...)
		return "", diags
	}

	return fp, diags
}
