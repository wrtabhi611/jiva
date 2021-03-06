/*
 Copyright © 2020 The OpenEBS Authors

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

package main

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

func getJivaImageID() string {
	var jivaImageID string
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if strings.Contains(image.RepoTags[0], "openebs/jiva") {
			logrus.Infof("Image: %v", image)
			jivaImageID = image.ID
			break
		}
	}
	return jivaImageID
}

func getJivaDebugImageID() string {
	var jivaImageID string
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if strings.Contains(image.RepoTags[0], "openebs/jiva") &&
			strings.Contains(image.RepoTags[0], "DEBUG") {
			jivaImageID = image.ID
		}
	}
	return jivaImageID
}

func createReplica(replicaIP string, config *testConfig) string {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: config.Image,
			Cmd: func() []string {
				var envs []string
				for key, value := range config.ReplicaEnvs {
					pair := []string{key, value}
					envs = append(envs, pair...)
				}
				args := []string{
					"launch", "replica",
					"--frontendIP", config.ControllerIP,
					"--listen", replicaIP + ":9502",
					"--size", "5g", "/vol",
				}
				return append(envs, args...)
			}(),
			ExposedPorts: nat.PortSet{
				"9502/tcp": {},
				"9503/tcp": {},
				"9504/tcp": {},
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 0,
			},
			NetworkMode:     "stg-net",
			PublishAllPorts: true,
			Binds:           []string{"/tmp1/" + replicaIP + "vol:/vol"},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"stg-net": &network.EndpointSettings{
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: replicaIP,
					},
				},
			},
		}, "Replica_"+replicaIP,
	)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return resp.ID
}

func createController(controllerIP string, config *testConfig) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: config.Image,
			Cmd:   []string{"env", "REPLICATION_FACTOR=" + config.ReplicationFactor, "launch", "controller", "--frontend", "gotgt", "--frontendIP", controllerIP, config.VolumeName},
			ExposedPorts: nat.PortSet{
				"3260/tcp": {},
				"9501/tcp": {},
			},
		},
		&container.HostConfig{
			NetworkMode:     "stg-net",
			PublishAllPorts: true,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"stg-net": &network.EndpointSettings{
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: controllerIP,
					},
				},
			},
		}, "controller_"+controllerIP,
	)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	config.Controller[controllerIP] = resp.ID
}

func deleteController(config *testConfig) {
	stopContainer(config.Controller[config.ControllerIP])
	removeContainer(config.Controller[config.ControllerIP])
}

func stopContainer(containerID string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStop(ctx, containerID, nil); err != nil {
		panic(err)
	}
}

func removeContainer(containerID string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		panic(err)
	}
}

func startContainer(containerID string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func verifyRestartCount(containerID string, restartCount int) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containerInspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}
	for {
		if containerInspect.ContainerJSONBase.RestartCount >= restartCount {
			break
		}
		time.Sleep(5 * time.Second)
	}
}
