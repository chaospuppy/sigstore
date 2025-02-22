//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kms

import (
	"context"
	"crypto"
	"fmt"
	"strings"

	"github.com/sigstore/sigstore/pkg/signature"
)

type providerInit func(context.Context, string, crypto.Hash, ...signature.RPCOption) (SignerVerifier, error)

type providers struct {
	providers map[string]providerInit
}

// AddProvider adds the provider implementation into the local cache
func AddProvider(keyResourceID string, init providerInit) {
	providersMux.providers[keyResourceID] = init
}

var providersMux = &providers{
	providers: map[string]providerInit{},
}

// Get returns a KMS SignerVerifier for the given resource string and hash function
func Get(ctx context.Context, keyResourceID string, hashFunc crypto.Hash, opts ...signature.RPCOption) (SignerVerifier, error) {
	for ref, providerInit := range providersMux.providers {
		if strings.HasPrefix(keyResourceID, ref) {
			return providerInit(ctx, keyResourceID, hashFunc, opts...)
		}
	}
	return nil, fmt.Errorf("no provider found for that key reference")
}

// SupportedProviders returns list of initialized providers
func SupportedProviders() []string {
	var keys []string
	for key := range providersMux.providers {
		keys = append(keys, key)
	}
	return keys
}

// SignerVerifier creates and verifies digital signatures over a message using a KMS service
type SignerVerifier interface {
	signature.SignerVerifier
	CreateKey(ctx context.Context, algorithm string) (crypto.PublicKey, error)
	CryptoSigner(ctx context.Context, errFunc func(error)) (crypto.Signer, crypto.SignerOpts, error)
	SupportedAlgorithms() []string
	DefaultAlgorithm() string
}
