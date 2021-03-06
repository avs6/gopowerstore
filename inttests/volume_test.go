/*
 *
 * Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package inttests

import (
	"context"
	"github.com/dell/gopowerstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

const TestVolumePrefix = "test_vol_"
const DefaultVolSize int64 = 1048576

func createVol(t *testing.T) (string, string) {
	volName := TestVolumePrefix + randString(8)
	createParams := gopowerstore.VolumeCreate{}
	createParams.Name = &volName
	size := DefaultVolSize
	createParams.Size = &size
	createResp, err := C.CreateVolume(context.Background(), &createParams)
	checkAPIErr(t, err)
	return createResp.ID, volName
}

func deleteVol(t *testing.T, id string) {
	_, err := C.DeleteVolume(context.Background(), nil, id)
	checkAPIErr(t, err)
}

func createSnap(volID string, t *testing.T, volName string) gopowerstore.CreateResponse {
	volume, err := C.GetVolume(context.Background(), volID)
	checkAPIErr(t, err)
	assert.NotEmpty(t, volume.Name)
	assert.Equal(t, volName, volume.Name)
	snapName := volName + "_snapshot"
	snapDesc := "just a description"
	snap, snapCreateErr := C.CreateSnapshot(context.Background(), &gopowerstore.SnapshotCreate{
		Name:        &snapName,
		Description: &snapDesc,
	}, volID)
	checkAPIErr(t, snapCreateErr)
	return snap
}

func TestGetSnapshotsByVolumeID(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)

	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)

	snapList, err := C.GetSnapshotsByVolumeID(context.Background(), volID)
	checkAPIErr(t, err)

	assert.Equal(t, 1, len(snapList))
	assert.Equal(t, snap.ID, snapList[0].ID)
}

func TestGetSnapshot(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)

	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)

	got, err := C.GetSnapshot(context.Background(), snap.ID)
	checkAPIErr(t, err)

	assert.Equal(t, snap.ID, got.ID)
}

func TestGetSnapshots(t *testing.T) {
	_, err := C.GetSnapshots(context.Background())
	checkAPIErr(t, err)
}

func TestGetNonExistingSnapshot(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)

	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)
	_, err := C.DeleteSnapshot(context.Background(), nil, snap.ID)
	assert.NotEmpty(t, snap.ID)

	got, err := C.GetSnapshot(context.Background(), snap.ID)
	assert.Error(t, err)
	assert.Empty(t, got)
}

func TestCreateSnapshot(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)
	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)
}

func TestDeleteSnapshot(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)
	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)
	_, err := C.DeleteSnapshot(context.Background(), nil, snap.ID)
	checkAPIErr(t, err)
}

func TestCreateVolumeFromSnapshot(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)
	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)

	name := "new_volume_from_snap" + randString(8)
	createParams := gopowerstore.VolumeClone{}
	createParams.Name = &name
	snapVol, err := C.CreateVolumeFromSnapshot(context.Background(), &createParams, snap.ID)
	checkAPIErr(t, err)
	assert.NotEmpty(t, snapVol.ID)
	deleteVol(t, snapVol.ID)
}

func TestGetVolumes(t *testing.T) {
	_, err := C.GetVolumes(context.Background())
	checkAPIErr(t, err)
}

func TestGetVolume(t *testing.T) {
	volID, volName := createVol(t)
	volume, err := C.GetVolume(context.Background(), volID)
	checkAPIErr(t, err)
	assert.NotEmpty(t, volume.Name)
	assert.Equal(t, volName, volume.Name)
	deleteVol(t, volID)
}

func TestGetVolumeByName(t *testing.T) {
	volID, volName := createVol(t)
	volume, err := C.GetVolumeByName(context.Background(), volName)
	checkAPIErr(t, err)
	assert.NotEmpty(t, volume.Name)
	assert.Equal(t, volName, volume.Name)
	deleteVol(t, volID)
}

func TestCreateDeleteVolume(t *testing.T) {
	volID, _ := createVol(t)
	deleteVol(t, volID)
}

func TestDeleteUnknownVol(t *testing.T) {
	volID := "f98de58e-9223-4fdc-86bd-d4ff268e20e1"
	_, err := C.DeleteVolume(context.Background(), nil, volID)
	if err != nil {
		apiError, ok := err.(gopowerstore.APIError)
		if !ok {
			t.Log("Unexpected API response")
			t.FailNow()
		}
		assert.True(t, apiError.VolumeIsNotExist())
	}
}

func TestGetVolumesWithTrace(t *testing.T) {
	ctx := C.SetTraceID(context.Background(),
		"126c9213-11d4-40b4-8da2-8cd70e277fe4")
	_, err := C.GetVolumes(ctx)
	checkAPIErr(t, err)
}

func TestVolumeAlreadyExist(t *testing.T) {
	volID, name := createVol(t)
	defer deleteVol(t, volID)
	createReq := gopowerstore.VolumeCreate{}
	createReq.Name = &name
	size := DefaultVolSize
	createReq.Size = &size
	_, err := C.CreateVolume(context.Background(), &createReq)
	assert.NotNil(t, err)
	apiError := err.(gopowerstore.APIError)
	assert.True(t, apiError.VolumeNameIsAlreadyUse())
}

func TestSnapshotAlreadyExist(t *testing.T) {
	volID, volName := createVol(t)
	defer deleteVol(t, volID)
	snap := createSnap(volID, t, volName)
	assert.NotEmpty(t, snap.ID)

	snapName := volName + "_snapshot"
	snapDesc := "just a description"
	snap, err := C.CreateSnapshot(context.Background(), &gopowerstore.SnapshotCreate{
		Name:        &snapName,
		Description: &snapDesc,
	}, volID)
	assert.NotNil(t, err)
	apiError := err.(gopowerstore.APIError)
	assert.True(t, apiError.SnapshotNameIsAlreadyUse())
}

func TestGetInvalidVolume(t *testing.T) {
	_, err := C.GetVolume(context.Background(), "4961282c-c5c5-4234-935f-2742fed499d0")
	assert.NotNil(t, err)
	apiError := err.(gopowerstore.APIError)
	assert.True(t, apiError.VolumeIsNotExist())
}
