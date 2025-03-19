// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/openconfig/containerz/client"
)

var (
	cntCommand, instance string
	ports                []string
	envs                 []string
	volumes              []string
	devices              []string
	network              string
	runAs                string
	restartPolicy        string
	addCaps              []string
	delCaps              []string
	cpus                 float64
	softMem              int64
	hardMem              int64
)

var cntStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start a container wirth the specified image and tag",
	RunE: func(command *cobra.Command, args []string) error {
		if image == "" {
			return fmt.Errorf("--image must be specified")
		}

		opts := []client.StartOption{}
		if len(ports) > 0 {
			opts = append(opts, client.WithPorts(ports))
		}
		if len(envs) > 0 {
			opts = append(opts, client.WithEnv(envs))
		}
		if len(volumes) > 0 {
			opts = append(opts, client.WithVolumes(volumes))
		}
		if len(devices) > 0 {
			opts = append(opts, client.WithDevices(devices))
		}
		if network != "" {
			opts = append(opts, client.WithNetwork(network))
		}
		if runAs != "" {
			opts = append(opts, client.WithRunAs(runAs))
		}
		if restartPolicy != "" {
			opts = append(opts, client.WithRestartPolicy(restartPolicy))
		}
		if len(addCaps) > 0 || len(delCaps) > 0 {
			opts = append(opts, client.WithCapabilities(addCaps, delCaps))
		}
		if len(labels) > 0 {
			label := map[string]string{}
			for _, l := range labels {
				parts := strings.SplitN(l, "=", 2)
				label[parts[0]] = parts[1]
			}
			opts = append(opts, client.WithLabels(label))
		}
		if cpus > 0 {
			opts = append(opts, client.WithCPUs(cpus))
		}
		if softMem > 0 {
			opts = append(opts, client.WithSoftLimit(softMem))
		}
		if hardMem > 0 {
			opts = append(opts, client.WithHardLimit(hardMem))
		}

		id, err := containerzClient.StartContainer(command.Context(), image, tag, cntCommand, instance, opts...)
		if err != nil {
			return err
		}

		fmt.Printf("Container started with id - %s\n", id)
		return nil
	},
}

func init() {
	containerCmd.AddCommand(cntStartCmd)

	cntStartCmd.PersistentFlags().StringVar(&cntCommand, "command", "/bin/bash", "command to run.")
	cntStartCmd.PersistentFlags().StringVar(&instance, "instance", "", "Name to give to the container.")
	cntStartCmd.PersistentFlags().StringVar(&network, "network", "", "Network to attach container to.")
	cntStartCmd.PersistentFlags().StringVar(&runAs, "runas", "", "User to use (format: <user>[:<group>]")
	cntStartCmd.PersistentFlags().StringVar(&restartPolicy, "restart_policy", "", "Restart policy to use. "+
		"Valid policies are \"always\", \"on-failure\", \"unless-stopped\", and \"none\". "+
		"Some policies (e.g., \"on-failure\") optionally accept a maximum number of restart attempts. "+
		"(format: <policy>[:<max_attempts>])")
	cntStartCmd.PersistentFlags().StringArrayVar(&ports, "port", []string{}, "Ports to expose (format: <internal_port>:<external_port>")
	cntStartCmd.PersistentFlags().StringArrayVar(&envs, "env", []string{}, "Environment vars to set (format: <VAR_NAMEt>=<VAR_VALUE>")
	cntStartCmd.PersistentFlags().StringArrayVarP(&volumes, "volume", "v", []string{}, "Volumes to attach to the container (format: <volume-name>:<mountpoint>[:ro])")
	cntStartCmd.PersistentFlags().StringArrayVarP(&devices, "device", "d", []string{}, "Devices to attach to the container (format: <src-path>[:<dst-path>[:<permissions>]])")
	cntStartCmd.PersistentFlags().StringArrayVar(&addCaps, "add_caps", []string{}, "Capabilities to add.")
	cntStartCmd.PersistentFlags().StringArrayVar(&delCaps, "del_caps", []string{}, "Capabilities to remove.")
	cntStartCmd.PersistentFlags().StringArrayVar(&labels, "labels", []string{}, "Labels to add to the container (format: <key>=<value>).")
	cntStartCmd.PersistentFlags().Float64Var(&cpus, "cpus", 0.0, "CPU limit to set.")
	cntStartCmd.PersistentFlags().Int64Var(&softMem, "soft_mem", 0, "Soft memory limit to set.")
	cntStartCmd.PersistentFlags().Int64Var(&hardMem, "hard_mem", 0, "Hard memory limit to set.")
}
