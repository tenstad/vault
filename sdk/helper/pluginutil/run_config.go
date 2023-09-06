// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pluginutil

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-secure-stdlib/plugincontainer"
	"github.com/hashicorp/vault/sdk/helper/consts"
	"github.com/hashicorp/vault/sdk/helper/pluginruntimeutil"
)

type PluginClientConfig struct {
	Name            string
	PluginType      consts.PluginType
	Version         string
	PluginSets      map[int]plugin.PluginSet
	HandshakeConfig plugin.HandshakeConfig
	Logger          log.Logger
	IsMetadataMode  bool
	AutoMTLS        bool
	MLock           bool
	Wrapper         RunnerUtil
}

type runConfig struct {
	// Provided by PluginRunner
	command  string
	image    string
	imageTag string
	args     []string
	sha256   []byte

	// Initialized with what's in PluginRunner.Env, but can be added to
	env []string

	runtimeConfig *pluginruntimeutil.PluginRuntimeConfig

	PluginClientConfig
}

func (rc runConfig) makeConfig(ctx context.Context) (*plugin.ClientConfig, error) {
	cmd := exec.Command(rc.command, rc.args...)
	cmd.Env = append(cmd.Env, rc.env...)

	// Add the mlock setting to the ENV of the plugin
	if rc.MLock || (rc.Wrapper != nil && rc.Wrapper.MlockEnabled()) {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", PluginMlockEnabled, "true"))
	}
	version, err := rc.Wrapper.VaultVersion(ctx)
	if err != nil {
		return nil, err
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", PluginVaultVersionEnv, version))

	if rc.IsMetadataMode {
		rc.Logger = rc.Logger.With("metadata", "true")
	}
	metadataEnv := fmt.Sprintf("%s=%t", PluginMetadataModeEnv, rc.IsMetadataMode)
	cmd.Env = append(cmd.Env, metadataEnv)

	automtlsEnv := fmt.Sprintf("%s=%t", PluginAutoMTLSEnv, rc.AutoMTLS)
	cmd.Env = append(cmd.Env, automtlsEnv)

	var clientTLSConfig *tls.Config
	if !rc.AutoMTLS && !rc.IsMetadataMode {
		// Get a CA TLS Certificate
		certBytes, key, err := generateCert()
		if err != nil {
			return nil, err
		}

		// Use CA to sign a client cert and return a configured TLS config
		clientTLSConfig, err = createClientTLSConfig(certBytes, key)
		if err != nil {
			return nil, err
		}

		// Use CA to sign a server cert and wrap the values in a response wrapped
		// token.
		wrapToken, err := wrapServerConfig(ctx, rc.Wrapper, certBytes, key)
		if err != nil {
			return nil, err
		}

		// Add the response wrap token to the ENV of the plugin
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", PluginUnwrapTokenEnv, wrapToken))
	}

	clientConfig := &plugin.ClientConfig{
		HandshakeConfig:  rc.HandshakeConfig,
		VersionedPlugins: rc.PluginSets,
		TLSConfig:        clientTLSConfig,
		Logger:           rc.Logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
			plugin.ProtocolGRPC,
		},
		AutoMTLS: rc.AutoMTLS,
	}
	if rc.image == "" {
		clientConfig.Cmd = cmd
		clientConfig.SecureConfig = &plugin.SecureConfig{
			Checksum: rc.sha256,
			Hash:     sha256.New(),
		}
	} else {
		clientConfig.SkipHostEnv = true
		cfg := &plugincontainer.Config{
			Image:  rc.image,
			Tag:    rc.imageTag,
			SHA256: fmt.Sprintf("%x", rc.sha256),

			Env:      cmd.Env,
			GroupAdd: os.Getgid(),
			Runtime:  consts.DefaultContainerPluginOCIRuntime,
			Labels: map[string]string{
				"managed-by": "hashicorp.com/vault",
			},
		}
		if cmd.Path != "" {
			cfg.Entrypoint = []string{cmd.Path}
		}
		if len(cmd.Args) > 0 {
			cfg.Args = cmd.Args
		}
		if rc.runtimeConfig != nil {
			cfg.CgroupParent = rc.runtimeConfig.CgroupParent
			cfg.NanoCpus = rc.runtimeConfig.CPU
			cfg.Memory = rc.runtimeConfig.Memory
			if rc.runtimeConfig.OCIRuntime != "" {
				cfg.Runtime = rc.runtimeConfig.OCIRuntime
			}
		}
		clientConfig.RunnerFunc = cfg.NewContainerRunner
		clientConfig.UnixSocketConfig = &plugin.UnixSocketConfig{
			Group: strconv.Itoa(os.Getgid()),
		}
	}
	return clientConfig, nil
}

func (rc runConfig) run(ctx context.Context) (*plugin.Client, error) {
	clientConfig, err := rc.makeConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := plugin.NewClient(clientConfig)
	return client, nil
}

type RunOpt func(*runConfig)

func Env(env ...string) RunOpt {
	return func(rc *runConfig) {
		rc.env = append(rc.env, env...)
	}
}

func Runner(wrapper RunnerUtil) RunOpt {
	return func(rc *runConfig) {
		rc.Wrapper = wrapper
	}
}

func PluginSets(pluginSets map[int]plugin.PluginSet) RunOpt {
	return func(rc *runConfig) {
		rc.PluginSets = pluginSets
	}
}

func HandshakeConfig(hs plugin.HandshakeConfig) RunOpt {
	return func(rc *runConfig) {
		rc.HandshakeConfig = hs
	}
}

func Logger(logger log.Logger) RunOpt {
	return func(rc *runConfig) {
		rc.Logger = logger
	}
}

func MetadataMode(isMetadataMode bool) RunOpt {
	return func(rc *runConfig) {
		rc.IsMetadataMode = isMetadataMode
	}
}

func AutoMTLS(autoMTLS bool) RunOpt {
	return func(rc *runConfig) {
		rc.AutoMTLS = autoMTLS
	}
}

func MLock(mlock bool) RunOpt {
	return func(rc *runConfig) {
		rc.MLock = mlock
	}
}

func (r *PluginRunner) RunConfig(ctx context.Context, opts ...RunOpt) (*plugin.Client, error) {
	var image, imageTag string
	if r.OCIImage != "" {
		image = r.OCIImage
		imageTag = strings.TrimPrefix(r.Version, "v")
	}
	rc := runConfig{
		command:       r.Command,
		image:         image,
		imageTag:      imageTag,
		args:          r.Args,
		sha256:        r.Sha256,
		env:           r.Env,
		runtimeConfig: r.RuntimeConfig,
	}

	for _, opt := range opts {
		opt(&rc)
	}

	return rc.run(ctx)
}
