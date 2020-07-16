// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

package linter

import (
	"path"
	"testing"

	"github.com/primait/drone-yaml/yaml"
)

func TestLint(t *testing.T) {
	tests := []struct {
		path    string
		trusted bool
		invalid bool
		message string
	}{
		{
			path:    "testdata/simple.yml",
			trusted: false,
			invalid: false,
		},
		{
			path:    "testdata/invalid_os.yml",
			trusted: false,
			invalid: true,
			message: "linter: unsupported os: openbsd",
		},
		{
			path:    "testdata/invalid_arch.yml",
			trusted: false,
			invalid: true,
			message: "linter: unsupported architecture: s390x",
		},
		{
			path:    "testdata/missing_build_image.yml",
			invalid: true,
			message: "linter: invalid or missing build image",
		},
		{
			path:    "testdata/missing_image.yml",
			invalid: true,
			message: "linter: invalid or missing image",
		},
		{
			path:    "testdata/missing_name.yml",
			invalid: true,
			message: "linter: invalid or missing name",
		},
		// user should not use reserved volume names.
		{
			path:    "testdata/volume_invalid_name.yml",
			trusted: false,
			invalid: true,
			message: "linter: invalid volume name: _workspace",
		},
		{
			path:    "testdata/pipeline_volume_invalid_name.yml",
			trusted: false,
			invalid: true,
			message: "linter: invalid volume name: _docker_socket",
		},
		// user should not be able to mount host path
		// volumes unless the repository is trusted.
		{
			path:    "testdata/volume_host_path.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount host volumes",
		},
		{
			path:    "testdata/volume_host_path.yml",
			trusted: true,
			invalid: false,
		},
		// user should be able to mount emptyDir volumes
		// where no medium is specified.
		{
			path:    "testdata/volume_empty_dir.yml",
			trusted: false,
			invalid: false,
		},
		// user should not be able to mount in-memory
		// emptyDir volumes unless the repository is
		// trusted.
		{
			path:    "testdata/volume_empty_dir_memory.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount in-memory volumes",
		},
		{
			path:    "testdata/volume_empty_dir_memory.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be able to bind to host ports
		// or IP addresses unless the repository is trusted.
		{
			path:    "testdata/service_port_host.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot map to a host port",
		},
		{
			path:    "testdata/service_port_host.yml",
			trusted: true,
			invalid: false,
		},
		{
			path:    "testdata/pipeline_port_host.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot map to a host port",
		},
		{
			path:    "testdata/pipeline_port_host.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be able to mount devices unless
		// the repository is trusted.
		{
			path:    "testdata/service_device.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount devices",
		},
		{
			path:    "testdata/service_device.yml",
			trusted: true,
			invalid: false,
		},
		{
			path:    "testdata/pipeline_device.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot mount devices",
		},
		{
			path:    "testdata/pipeline_device.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be able to set the securityContext
		// unless the repository is trusted.
		{
			path:    "testdata/pipeline_privileged.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot enable privileged mode",
		},
		{
			path:    "testdata/pipeline_privileged.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be able to set dns, dns_search or
		// extra_hosts unless the repository is trusted.
		{
			path:    "testdata/pipeline_dns.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot configure dns",
		},
		{
			path:    "testdata/pipeline_dns.yml",
			trusted: true,
			invalid: false,
		},
		{
			path:    "testdata/pipeline_dns_search.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot configure dns_search",
		},
		{
			path:    "testdata/pipeline_dns_search.yml",
			trusted: true,
			invalid: false,
		},
		{
			path:    "testdata/pipeline_extra_hosts.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot configure extra_hosts",
		},
		{
			path:    "testdata/pipeline_extra_hosts.yml",
			trusted: true,
			invalid: false,
		},
		{
			path:    "testdata/pipeline_network_mode.yml",
			trusted: false,
			invalid: true,
			message: "linter: untrusted repositories cannot configure network_mode",
		},
		{
			path:    "testdata/pipeline_network_mode.yml",
			trusted: true,
			invalid: false,
		},
		// user should not be able to use duplicate names
		// for steps or services.
		{
			path:    "testdata/duplicate_step.yml",
			invalid: true,
			message: "linter: duplicate step names",
		},
		{
			path:    "testdata/duplicate_step_service.yml",
			invalid: true,
			message: "linter: duplicate step names",
		},
	}
	for _, test := range tests {
		name := path.Base(test.path)
		if test.trusted {
			name = name + "/trusted"
		}
		t.Run(name, func(t *testing.T) {
			resources, err := yaml.ParseFile(test.path)
			if err != nil {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Error(err)
				return
			}

			err = Lint(resources.Resources[0], test.trusted)
			if err == nil && test.invalid == true {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Expect lint error")
				return
			}

			if err != nil && test.invalid == false {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Expect lint error is nil, got %s", err)
				return
			}

			if err == nil {
				return
			}

			if got, want := err.Error(), test.message; got != want {
				t.Logf("yaml: %s", test.path)
				t.Logf("trusted: %v", test.trusted)
				t.Errorf("Want message %q, got %q", want, got)
				return
			}
		})
	}
}
