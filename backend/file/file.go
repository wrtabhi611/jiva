/*
 Copyright © 2020 The OpenEBS Authors

 This file was originally authored by Rancher Labs
 under Apache License 2018.

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

package file

import (
	"os"

	"github.com/openebs/jiva/types"
	"github.com/sirupsen/logrus"
)

func New() types.BackendFactory {
	return &Factory{}
}

type Factory struct {
}

type Wrapper struct {
	*os.File
}

func (f *Wrapper) Close() error {
	logrus.Infof("Closing: %s", f.Name())
	return f.File.Close()
}

func (f *Wrapper) Snapshot(name string, userCreated bool, created string) error {
	return nil
}

func (f *Wrapper) Resize(name string, size string) error {
	return nil
}

func (f *Wrapper) Size() (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func (f *Wrapper) SectorSize() (int64, error) {
	return 4096, nil
}

func (f *Wrapper) RemainSnapshots() (int, error) {
	return 1, nil
}

func (f *Wrapper) GetRevisionCounter() (int64, error) {
	return 1, nil
}

func (f *Wrapper) GetVolUsage() (types.VolUsage, error) {
	return types.VolUsage{}, nil
}

// SetReplicaMode ...
func (f *Wrapper) SetReplicaMode(mode types.Mode) error {
	return nil
}

// SetCheckpoint ...
func (f *Wrapper) SetCheckpoint(snapshotName string) error {
	return nil
}

// GetReplicaChain ...
func (f *Wrapper) GetReplicaChain() ([]string, error) {
	return nil, nil
}

func (f *Wrapper) SetRevisionCounter(counter int64) error {
	return nil
}

func (f *Wrapper) SetRebuilding(rebuilding bool) error {
	return nil
}

func (ff *Factory) Create(address string) (types.Backend, error) {
	logrus.Infof("Creating file: %s", address)
	file, err := os.OpenFile(address, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		logrus.Infof("Failed to create file %s: %v", address, err)
		return nil, err
	}

	return &Wrapper{file}, nil
}

func (ff *Factory) SignalToAdd(address string, action string) error {
	return nil
}

func (f *Wrapper) GetMonitorChannel() types.MonitorChannel {
	return nil
}

func (f *Wrapper) GetCloneStatus() (string, error) {
	return "", nil
}

func (f *Wrapper) PingResponse() error {
	return nil
}

func (f *Wrapper) StopMonitoring() {
}

func (f *Wrapper) Sync() (int, error) {
	return 0, nil
}

func (f *Wrapper) Unmap(offset int64, length int64) (int, error) {
	return 0, nil
}
