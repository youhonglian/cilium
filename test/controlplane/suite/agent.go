// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package suite

import (
	"context"
	"fmt"
	"os"

	"github.com/cilium/cilium/daemon/cmd"
	fakeDatapath "github.com/cilium/cilium/pkg/datapath/fake"
	"github.com/cilium/cilium/pkg/endpoint"
	k8sClient "github.com/cilium/cilium/pkg/k8s/client"
	agentOption "github.com/cilium/cilium/pkg/option"
)

type agentHandle struct {
	d       *cmd.Daemon
	cancel  context.CancelFunc
	clean   func()
	tempDir string
}

func (h *agentHandle) tearDown() {
	h.d.Close()
	h.cancel()
	h.clean()
	os.RemoveAll(h.tempDir)
}

func startCiliumAgent(nodeName string, clientset k8sClient.Clientset) (*fakeDatapath.FakeDatapath, agentHandle, error) {
	var handle agentHandle

	handle.tempDir = setupTestDirectories()

	fdp := fakeDatapath.NewDatapath()

	ctx, cancel := context.WithCancel(context.Background())
	handle.cancel = cancel

	cleaner := cmd.NewDaemonCleanup()
	handle.clean = cleaner.Clean

	var err error
	handle.d, _, err = cmd.NewDaemon(ctx,
		cleaner,
		cmd.WithCustomEndpointManager(&dummyEpSyncher{}),
		fdp,
		clientset)
	if err != nil {
		return nil, agentHandle{}, err
	}
	return fdp, handle, nil
}

type dummyEpSyncher struct{}

func (epSync *dummyEpSyncher) RunK8sCiliumEndpointSync(e *endpoint.Endpoint, conf endpoint.EndpointStatusConfiguration) {
}

func (epSync *dummyEpSyncher) DeleteK8sCiliumEndpointSync(e *endpoint.Endpoint) {
}

func setupTestDirectories() string {
	tempDir, err := os.MkdirTemp("", "cilium-test-")
	if err != nil {
		panic(fmt.Sprintf("TempDir() failed: %s", err))
	}
	agentOption.Config.RunDir = tempDir
	agentOption.Config.StateDir = tempDir
	return tempDir
}
