/*
Copyright 2024 Keyfactor

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ejbca

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	ejbcaclient "github.com/Keyfactor/ejbca-go-client-sdk/api/ejbca"
	"github.com/hashicorp/go-hclog"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	commonutil "github.com/spiffe/spire/pkg/common/util"
	"github.com/spiffe/spire/pkg/server/plugin/upstreamauthority"
	"github.com/spiffe/spire/test/clock"
	"github.com/spiffe/spire/test/plugintest"
	"github.com/spiffe/spire/test/spiretest"
	"github.com/spiffe/spire/test/testkey"
	"github.com/spiffe/spire/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

var (
	trustDomain = spiffeid.RequireTrustDomainFromString("example.org")
)

type fakeEjbcaAuthenticator struct {
	client *http.Client
}

// GetHTTPClient implements ejbcaclient.Authenticator
func (f *fakeEjbcaAuthenticator) GetHTTPClient() (*http.Client, error) {
	return f.client, nil
}

type fakeClientConfig struct {
	testServer *httptest.Server
}

func (f *fakeClientConfig) newFakeAuthenticator(_ *Config) (ejbcaclient.Authenticator, error) {
	return &fakeEjbcaAuthenticator{
		client: f.testServer.Client(),
	}, nil
}

func TestConfigure(t *testing.T) {
	rootCA, _, svidIssuingCA, svidIssuingCAKey := issueTestCertificates(t)

	caPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCA.Raw})
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: svidIssuingCA.Raw})

	keyByte, err := x509.MarshalECPrivateKey(svidIssuingCAKey)
	require.NoError(t, err)
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyByte})

	for i, tt := range []struct {
		name     string
		getEnv   getEnvFunc
		readFile readFileFunc
		config   string

		expectedgRPCCode      codes.Code
		expectedMessagePrefix string
	}{
		{
			name: "No Auth Method",
			config: `
            hostname = "ejbca.example.org"
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `,
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Hostname",
			config: fmt.Sprintf(`
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_cert = <<EOF
%s
EOF
                client_key = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, certPem, keyPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No CA Name",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_cert = <<EOF
%s
EOF
                client_key = <<EOF
%s
EOF
            }
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, certPem, keyPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No End Entity Profile Name",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_cert = <<EOF
%s
EOF
                client_key = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, certPem, keyPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Certificate Profile Name",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_cert = <<EOF
%s
EOF
                client_key = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, certPem, keyPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Client Cert",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_key = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, keyPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Client Key",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_cert = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, certPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Token URL",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Client ID",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "No Client Secret",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv:           os.Getenv,
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.InvalidArgument,
		},
		{
			name: "Token URL from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_OAUTH_TOKEN_URL" {
					return "https://dev.idp.com/oauth/token"
				}
				return ""
			},
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Client ID from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_OAUTH_CLIENT_ID" {
					return "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
				}
				return ""
			},
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Client Secret from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_OAUTH_CLIENT_SECRET" {
					return "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
				}
				return ""
			},
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Scopes from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_OAUTH_SCOPES" {
					return "read:certificates,write:certificates"
				}
				return ""
			},
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Audience from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_OAUTH_AUDIENCE" {
					return "https://ejbca.example.com"
				}
				return ""
			},
			readFile:         os.ReadFile,
			expectedgRPCCode: codes.OK,
		},
		{
			name: "CA Cert path from env",
			config: `
            hostname = "ejbca.example.org"
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `,
			getEnv: func(key string) string {
				if key == "EJBCA_CA_CERT_PATH" {
					return "/path/to/ca.crt"
				}
				return ""
			},
			readFile: func(key string) ([]byte, error) {
				if key == "/path/to/ca.crt" {
					return caPem, nil
				}
				return nil, errors.New("file not found")
			},
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Client Cert path from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            cert_auth {
                client_key = <<EOF
%s
EOF
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem, keyPem),
			getEnv: func(key string) string {
				if key == "EJBCA_CLIENT_CERT_PATH" {
					return "/path/to/cert.crt"
				}
				return ""
			},
			readFile: func(key string) ([]byte, error) {
				if key == "/path/to/cert.crt" {
					return certPem, nil
				}
				return nil, errors.New("file not found")
			},
			expectedgRPCCode: codes.OK,
		},
		{
			name: "Client Key path from env",
			config: fmt.Sprintf(`
            hostname = "ejbca.example.org"
			ca_cert = <<EOF
%s
EOF
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `, caPem),
			getEnv: func(key string) string {
				if key == "EJBCA_CLIENT_CERT_KEY_PATH" {
					return "/path/to/key.pem"
				}
				return ""
			},
			readFile: func(key string) ([]byte, error) {
				if key == "/path/to/key.pem" {
					return keyPem, nil
				}
				return nil, errors.New("file not found")
			},
			expectedgRPCCode: codes.OK,
		},
		{
			name: "CA, Client Cert, and Client Key path from env",
			config: `
            hostname = "ejbca.example.org"
            cert_auth {}
            oauth {
                token_url = "https://dev.idp.com/oauth/token"
                client_id = "fi3ElQUVoBBHyRNt4mpUxG9WY65AOCcJ"
                client_secret = "1EXHdD7Ikmmv0OkBoJZZtzOG5iAzvwdqBVuvquf-QEvL6fLrEG_heJHphtEXVj9H"
                scopes = "read:certificates,write:certificates"
                audience = "https://ejbca.example.com"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `,
			getEnv: func(key string) string {
				if key == "EJBCA_CA_CERT_PATH" {
					return "/path/to/ca.crt"
				}
				if key == "EJBCA_CLIENT_CERT_PATH" {
					return "/path/to/cert.crt"
				}
				if key == "EJBCA_CLIENT_CERT_KEY_PATH" {
					return "/path/to/key.pem"
				}
				return ""
			},
			readFile: func(key string) ([]byte, error) {
				if key == "/path/to/ca.crt" {
					return caPem, nil
				}
				if key == "/path/to/cert.crt" {
					return certPem, nil
				}
				if key == "/path/to/key.pem" {
					return keyPem, nil
				}
				return nil, errors.New("file not found")
			},
			expectedgRPCCode: codes.OK,
		},
		{
			name: "CA, Client Cert, and Client Key path from config",
			config: `
            hostname = "ejbca.example.org"
			ca_cert_path = "/path/to/ca.crt"
            cert_auth {
				client_cert_path = "/path/to/cert.crt"
				client_key_path = "/path/to/key.pem"
            }
            ca_name = "Fake-Sub-CA"
            end_entity_profile_name = "fakeSpireIntermediateCAEEP"
            certificate_profile_name = "fakeSubCACP"
            default_end_entity_name = "cn"
            account_binding_id = "spiffe://example.org/spire/agent/join_token/abcd"
            `,
			getEnv: os.Getenv,
			readFile: func(key string) ([]byte, error) {
				if key == "/path/to/ca.crt" {
					return caPem, nil
				}
				if key == "/path/to/cert.crt" {
					return certPem, nil
				}
				if key == "/path/to/key.pem" {
					return keyPem, nil
				}
				return nil, errors.New("file not found")
			},
			expectedgRPCCode: codes.OK,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			p := New()

			p.hooks.getEnv = tt.getEnv
			p.hooks.readFile = tt.readFile

			logOptions := hclog.DefaultOptions
			logOptions.Level = hclog.Debug
			p.SetLogger(hclog.Default())

			options := []plugintest.Option{
				plugintest.CaptureConfigureError(&err),
				plugintest.Configure(tt.config),
			}

			plugintest.Load(t, builtin(p), new(upstreamauthority.V1), options...)
			spiretest.RequireGRPCStatusHasPrefix(t, err, tt.expectedgRPCCode, tt.expectedMessagePrefix)
			t.Logf("\ntestcase[%d] and err:%+v\n", i, err)
		})
	}
}

func TestMintX509CAAndSubscribe(t *testing.T) {
	rootCA, intermediateCA, svidIssuingCA, _ := issueTestCertificates(t)

	for _, tt := range []struct {
		name string

		// Config
		certificateResponseFormat string
		ejbcaStatusCode           int

		// Request
		caName                 string
		endEntityProfileName   string
		certificateProfileName string
		endEntityName          string
		accountBindingID       string

		// Expected values
		expectedgRPCCode      codes.Code
		expectedMessagePrefix string
		expectedEndEntityName string
		expectedCaAndChain    []*x509.Certificate
		expectedRootCAs       []*x509.Certificate
	}{
		{
			name: "success_pem",

			certificateResponseFormat: "PEM",
			ejbcaStatusCode:           http.StatusOK,

			caName:                 "Fake-Sub-CA",
			endEntityProfileName:   "fakeSpireIntermediateCAEEP",
			certificateProfileName: "fakeSubCACP",
			endEntityName:          "",
			accountBindingID:       "",

			expectedgRPCCode:      codes.OK,
			expectedMessagePrefix: "",
			expectedEndEntityName: trustDomain.ID().String(),
			expectedCaAndChain:    []*x509.Certificate{svidIssuingCA, intermediateCA},
			expectedRootCAs:       []*x509.Certificate{rootCA},
		},
		{
			name: "success_der",

			certificateResponseFormat: "DER",
			ejbcaStatusCode:           http.StatusOK,

			caName:                 "Fake-Sub-CA",
			endEntityProfileName:   "fakeSpireIntermediateCAEEP",
			certificateProfileName: "fakeSubCACP",
			endEntityName:          "",
			accountBindingID:       "",

			expectedgRPCCode:      codes.OK,
			expectedMessagePrefix: "",
			expectedEndEntityName: trustDomain.ID().String(),
			expectedCaAndChain:    []*x509.Certificate{svidIssuingCA, intermediateCA},
			expectedRootCAs:       []*x509.Certificate{rootCA},
		},
		{
			name: "fail_unknown_format",

			certificateResponseFormat: "PKCS7",

			caName:                 "Fake-Sub-CA",
			endEntityProfileName:   "fakeSpireIntermediateCAEEP",
			certificateProfileName: "fakeSubCACP",
			endEntityName:          "",
			accountBindingID:       "",

			expectedgRPCCode:      codes.Internal,
			expectedMessagePrefix: "upstreamauthority(ejbca): ejbca returned unsupported certificate format: PKCS7",
			ejbcaStatusCode:       http.StatusOK,
			expectedEndEntityName: trustDomain.ID().String(),
			expectedCaAndChain:    []*x509.Certificate{svidIssuingCA, intermediateCA},
			expectedRootCAs:       []*x509.Certificate{rootCA},
		},
		{
			name: "success_ejbca_api_error",

			certificateResponseFormat: "PEM",
			ejbcaStatusCode:           http.StatusBadRequest,

			caName:                 "Fake-Sub-CA",
			endEntityProfileName:   "fakeSpireIntermediateCAEEP",
			certificateProfileName: "fakeSubCACP",
			endEntityName:          "",
			accountBindingID:       "",

			expectedgRPCCode:      codes.Internal,
			expectedMessagePrefix: "upstreamauthority(ejbca): EJBCA returned an error: failed to enroll CSR - 400 Bad Request - EJBCA API returned error",
			expectedEndEntityName: trustDomain.ID().String(),
			expectedCaAndChain:    []*x509.Certificate{svidIssuingCA, intermediateCA},
			expectedRootCAs:       []*x509.Certificate{rootCA},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			testServer := httptest.NewTLSServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					enrollRestRequest := ejbcaclient.EnrollCertificateRestRequest{}
					err := json.NewDecoder(r.Body).Decode(&enrollRestRequest)
					require.NoError(t, err)

					// Perform assertions before fake enrollment
					require.Equal(t, tt.caName, enrollRestRequest.GetCertificateAuthorityName())
					require.Equal(t, tt.endEntityProfileName, enrollRestRequest.GetEndEntityProfileName())
					require.Equal(t, tt.certificateProfileName, enrollRestRequest.GetCertificateProfileName())
					require.Equal(t, tt.accountBindingID, enrollRestRequest.GetAccountBindingId())
					require.Equal(t, tt.expectedEndEntityName, enrollRestRequest.GetUsername())

					response := certificateRestResponseFromExpectedCerts(t, tt.expectedCaAndChain, tt.expectedRootCAs, tt.certificateResponseFormat)

					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(tt.ejbcaStatusCode)
					err = json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				}))
			defer testServer.Close()

			p := New()
			ua := new(upstreamauthority.V1)

			logOptions := hclog.DefaultOptions
			logOptions.Level = hclog.Debug
			p.SetLogger(hclog.Default())

			clientConfig := fakeClientConfig{
				testServer: testServer,
			}
			p.hooks.newAuthenticator = clientConfig.newFakeAuthenticator

			config := &Config{
				Hostname: testServer.URL,

				// We populate the client cert & client key to random values since newFakeAuthenticator doesn't have
				// any built-in authentication.
				CertAuth: &CertAuthConfig{
					ClientCert: "BEGIN CERTIFICATE ... END CERTIFICATE",
					ClientKey:  "BEGIN RSA PRIVATE KEY ... END RSA PRIVATE KEY",
				},

				CAName:                 tt.caName,
				EndEntityProfileName:   tt.endEntityProfileName,
				CertificateProfileName: tt.certificateProfileName,
				DefaultEndEntityName:   tt.endEntityName,
				AccountBindingID:       tt.accountBindingID,
			}

			options := []plugintest.Option{
				plugintest.CaptureConfigureError(&err),
				plugintest.ConfigureJSON(config),
			}

			plugintest.Load(t, builtin(p), ua, options...)
			require.NoError(t, err)

			priv := testkey.NewEC384(t)
			csr, err := commonutil.MakeCSR(priv, trustDomain.ID())
			require.NoError(t, err)

			ctx := context.Background()
			caAndChain, rootCAs, stream, err := ua.MintX509CA(ctx, csr, 30*time.Second)
			spiretest.RequireGRPCStatusHasPrefix(t, err, tt.expectedgRPCCode, tt.expectedMessagePrefix)
			if tt.expectedgRPCCode == codes.OK {
				require.NotNil(t, stream)
				require.NotNil(t, caAndChain)
				require.NotNil(t, rootCAs)
			}
		})
	}
}

func certificateRestResponseFromExpectedCerts(t *testing.T, issuingCaAndChain []*x509.Certificate, rootCAs []*x509.Certificate, format string) *ejbcaclient.CertificateRestResponse {
	require.NotEqual(t, 0, len(issuingCaAndChain))
	var issuingCa string
	if format == "PEM" {
		issuingCa = string(pem.EncodeToMemory(&pem.Block{Bytes: issuingCaAndChain[0].Raw, Type: "CERTIFICATE"}))
	} else {
		issuingCa = base64.StdEncoding.EncodeToString(issuingCaAndChain[0].Raw)
	}

	var caChain []string
	if format == "PEM" {
		for _, cert := range issuingCaAndChain[1:] {
			caChain = append(caChain, string(pem.EncodeToMemory(&pem.Block{Bytes: cert.Raw, Type: "CERTIFICATE"})))
		}
		for _, cert := range rootCAs {
			caChain = append(caChain, string(pem.EncodeToMemory(&pem.Block{Bytes: cert.Raw, Type: "CERTIFICATE"})))
		}
	} else {
		for _, cert := range issuingCaAndChain[1:] {
			caChain = append(caChain, base64.StdEncoding.EncodeToString(cert.Raw))
		}
		for _, cert := range rootCAs {
			caChain = append(caChain, base64.StdEncoding.EncodeToString(cert.Raw))
		}
	}

	response := &ejbcaclient.CertificateRestResponse{}
	response.SetResponseFormat(format)
	response.SetCertificate(issuingCa)
	response.SetCertificateChain(caChain)
	return response
}

func TestGetEndEntityName(t *testing.T) {
	for _, tt := range []struct {
		name string

		defaultEndEntityName string

		subject  string
		dnsNames []string
		uris     []string
		ips      []string

		expectedEndEntityName string
	}{
		{
			name:                 "defaultEndEntityName unset use cn",
			defaultEndEntityName: "",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "purplecat.example.com",
		},
		{
			name:                 "defaultEndEntityName unset use dns",
			defaultEndEntityName: "",
			subject:              "",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "reddog.example.com",
		},
		{
			name:                 "defaultEndEntityName unset use uri",
			defaultEndEntityName: "",
			subject:              "",
			dnsNames:             []string{""},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "https://blueelephant.example.com",
		},
		{
			name:                 "defaultEndEntityName unset use ip",
			defaultEndEntityName: "",
			subject:              "",
			dnsNames:             []string{""},
			uris:                 []string{""},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "192.168.1.1",
		},
		{
			name:                 "defaultEndEntityName set use cn",
			defaultEndEntityName: "cn",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "purplecat.example.com",
		},
		{
			name:                 "defaultEndEntityName set use dns",
			defaultEndEntityName: "dns",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "reddog.example.com",
		},
		{
			name:                 "defaultEndEntityName set use uri",
			defaultEndEntityName: "uri",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "https://blueelephant.example.com",
		},
		{
			name:                 "defaultEndEntityName set use ip",
			defaultEndEntityName: "ip",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "192.168.1.1",
		},
		{
			name:                 "defaultEndEntityName set use custom",
			defaultEndEntityName: "aNonStandardValue",
			subject:              "CN=purplecat.example.com",
			dnsNames:             []string{"reddog.example.com"},
			uris:                 []string{"https://blueelephant.example.com"},
			ips:                  []string{"192.168.1.1"},

			expectedEndEntityName: "aNonStandardValue",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Hostname: "ejbca.example.com",
				CertAuth: &CertAuthConfig{
					ClientCert: "BEGIN CERTIFICATE ... END CERTIFICATE",
					ClientKey:  "BEGIN RSA PRIVATE KEY ... END RSA PRIVATE KEY",
				},
				CAName:                 "Fake-Sub-CA",
				EndEntityProfileName:   "fakeSpireIntermediateCAEEP",
				CertificateProfileName: "fakeSubCACP",
				DefaultEndEntityName:   tt.defaultEndEntityName,
				AccountBindingID:       "",
			}

			csr, err := generateCSR(tt.subject, tt.dnsNames, tt.uris, tt.ips)
			require.NoError(t, err)

			p := New()

			logOptions := hclog.DefaultOptions
			logOptions.Level = hclog.Debug
			p.SetLogger(hclog.Default())

			endEntityName, err := p.getEndEntityName(config, csr)
			require.NoError(t, err)
			require.Equal(t, tt.expectedEndEntityName, endEntityName)
		})
	}
}

func generateCSR(subject string, dnsNames []string, uris []string, ipAddresses []string) (*x509.CertificateRequest, error) {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 2048)

	var name pkix.Name

	if subject != "" {
		// Split the subject into its individual parts
		parts := strings.Split(subject, ",")

		for _, part := range parts {
			// Split the part into key and value
			keyValue := strings.SplitN(part, "=", 2)

			if len(keyValue) != 2 {
				return nil, errors.New("invalid subject")
			}

			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			// Map the key to the appropriate field in the pkix.Name struct
			switch key {
			case "C":
				name.Country = []string{value}
			case "ST":
				name.Province = []string{value}
			case "L":
				name.Locality = []string{value}
			case "O":
				name.Organization = []string{value}
			case "OU":
				name.OrganizationalUnit = []string{value}
			case "CN":
				name.CommonName = value
			default:
				// Ignore any unknown keys
			}
		}
	}

	template := x509.CertificateRequest{
		Subject:            name,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	if len(dnsNames) > 0 {
		template.DNSNames = dnsNames
	}

	// Parse and add URIs
	var uriPointers []*url.URL
	for _, u := range uris {
		if u == "" {
			continue
		}
		uriPointer, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		uriPointers = append(uriPointers, uriPointer)
	}
	template.URIs = uriPointers

	// Parse and add IPAddresses
	var ipAddrs []net.IP
	for _, ipStr := range ipAddresses {
		if ipStr == "" {
			continue
		}
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}
		ipAddrs = append(ipAddrs, ip)
	}
	template.IPAddresses = ipAddrs

	// Generate the CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		return nil, err
	}

	parsedCSR, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, err
	}

	return parsedCSR, nil
}

func issueTestCertificates(t *testing.T) (*x509.Certificate, *x509.Certificate, *x509.Certificate, *ecdsa.PrivateKey) {
	now := clock.NewMock(t).Now()
	rootCaTemplate := &x509.Certificate{
		Subject:               pkix.Name{CommonName: "Fake-Root-CA"},
		SerialNumber:          big.NewInt(1),
		BasicConstraintsValid: true,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(time.Hour * 24),
	}
	rootCA, rootCAKey, err := util.SelfSign(rootCaTemplate)
	require.NoError(t, err)

	intermediateCATemplate := &x509.Certificate{
		Subject:               pkix.Name{CommonName: "Fake-Sub-CA"},
		SerialNumber:          big.NewInt(1),
		BasicConstraintsValid: true,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(time.Hour * 24),
	}
	intermediateCA, intermediateKey, err := util.Sign(intermediateCATemplate, rootCA, rootCAKey)
	require.NoError(t, err)

	svidIssuingCATemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		BasicConstraintsValid: true,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(time.Hour * 24),
		URIs:                  []*url.URL{trustDomain.ID().URL()},
	}
	svidIssuingCA, svidIssuingCAKey, err := util.Sign(svidIssuingCATemplate, intermediateCA, intermediateKey)
	require.NoError(t, err)

	return rootCA, intermediateCA, svidIssuingCA, svidIssuingCAKey
}
