/*
 * Copyright The Kmesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package workload

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/rand"

	"kmesh.net/kmesh/api/v2/workloadapi"
	"kmesh.net/kmesh/daemon/options"
	"kmesh.net/kmesh/pkg/bpf"
	"kmesh.net/kmesh/pkg/constants"
	"kmesh.net/kmesh/pkg/controller/workload/bpfcache"
	"kmesh.net/kmesh/pkg/controller/workload/cache"
	"kmesh.net/kmesh/pkg/nets"
	"kmesh.net/kmesh/pkg/utils/test"
)

func Test_handleWorkload(t *testing.T) {
	workloadMap := bpfcache.NewFakeWorkloadMap(t)
	defer bpfcache.CleanupFakeWorkloadMap(workloadMap)

	p := newProcessor(workloadMap)

	// 1. handle workload with service, but service not handled yet
	// In this case, only frontend map and backend map should be updated.
	wl := createTestWorkloadWithService(true)
	err := p.handleWorkload(wl)
	assert.NoError(t, err)

	var (
		ek bpfcache.EndpointKey
		ev bpfcache.EndpointValue
	)

	workloadID := checkFrontEndMap(t, wl.Addresses[0], p)
	checkBackendMap(t, p, workloadID, wl)

	epKeys := p.bpf.GetEndpointKeys(workloadID)
	assert.Equal(t, len(epKeys), 0)
	for svcName := range wl.Services {
		endpoints := p.endpointsByService[svcName]
		assert.Len(t, endpoints, 1)
		if _, ok := endpoints[wl.Uid]; ok {
			assert.True(t, ok)
		}
	}

	// 2. add related service
	fakeSvc := createFakeService("testsvc", "10.240.10.1", "10.240.10.2")
	_ = p.handleService(fakeSvc)

	// 2.1 check front end map contains service
	svcID := checkFrontEndMap(t, fakeSvc.Addresses[0].Address, p)

	// 2.2 check service map contains service
	checkServiceMap(t, p, svcID, fakeSvc, 1)

	// 2.3 check endpoint map now contains the workloads
	ek.BackendIndex = 1
	ek.ServiceId = svcID
	err = p.bpf.EndpointLookup(&ek, &ev)
	assert.NoError(t, err)
	assert.Equal(t, ev.BackendUid, workloadID)

	// 3. add another workload with service
	workload2 := createFakeWorkload("1.2.3.5", workloadapi.NetworkMode_STANDARD)
	err = p.handleWorkload(workload2)
	assert.NoError(t, err)

	// 3.1 check endpoint map now contains the new workloads
	workload2ID := checkFrontEndMap(t, workload2.Addresses[0], p)
	ek.BackendIndex = 2
	ek.ServiceId = svcID
	err = p.bpf.EndpointLookup(&ek, &ev)
	assert.NoError(t, err)
	assert.Equal(t, ev.BackendUid, workload2ID)

	// 3.2 check service map contains service
	checkServiceMap(t, p, svcID, fakeSvc, 2)

	// 4 modify workload2 attribute not relationship with services
	workload2.Waypoint = &workloadapi.GatewayAddress{
		Destination: &workloadapi.GatewayAddress_Address{
			Address: &workloadapi.NetworkAddress{
				Address: netip.MustParseAddr("10.10.10.10").AsSlice(),
			},
		},
		HboneMtlsPort: 15008,
	}

	err = p.handleWorkload(workload2)
	assert.NoError(t, err)

	// 4.1 check endpoint map now contains the new workloads
	workload2ID = checkFrontEndMap(t, workload2.Addresses[0], p)
	ek.BackendIndex = 2
	ek.ServiceId = svcID
	err = p.bpf.EndpointLookup(&ek, &ev)
	assert.NoError(t, err)
	assert.Equal(t, ev.BackendUid, workload2ID)

	// 4.2 check service map contains service
	checkServiceMap(t, p, svcID, fakeSvc, 2)

	// 4.3 check backend map contains waypoint
	checkBackendMap(t, p, workload2ID, workload2)

	// 5 update workload to remove the bound services
	wl3 := proto.Clone(wl).(*workloadapi.Workload)
	wl3.Services = nil
	err = p.handleWorkload(wl3)
	assert.NoError(t, err)

	// 5.1 check service map
	checkServiceMap(t, p, svcID, fakeSvc, 1)

	// 5.2 check endpoint map
	ek.BackendIndex = 1
	ek.ServiceId = svcID
	err = p.bpf.EndpointLookup(&ek, &ev)
	assert.NoError(t, err)
	assert.Equal(t, workload2ID, ev.BackendUid)

	// 6. add namespace scoped waypoint service
	wpSvc := createFakeService("waypoint", "10.240.10.5", "10.240.10.5")
	_ = p.handleService(wpSvc)
	assert.Nil(t, wpSvc.Waypoint)
	// 6.1 check front end map contains service
	svcID = checkFrontEndMap(t, wpSvc.Addresses[0].Address, p)
	// 6.2 check service map contains service, but no waypoint address
	checkServiceMap(t, p, svcID, wpSvc, 0)

	hashNameClean(p)
}

func Test_hostnameNetworkMode(t *testing.T) {
	workloadMap := bpfcache.NewFakeWorkloadMap(t)
	p := newProcessor(workloadMap)
	workload := createFakeWorkload("1.2.3.4", workloadapi.NetworkMode_STANDARD)
	workloadWithoutService := createFakeWorkload("1.2.3.5", workloadapi.NetworkMode_STANDARD)
	workloadWithoutService.Services = nil
	workloadHostname := createFakeWorkload("1.2.3.6", workloadapi.NetworkMode_HOST_NETWORK)

	p.handleWorkload(workload)
	p.handleWorkload(workloadWithoutService)
	p.handleWorkload(workloadHostname)

	// Check Workload Cache
	checkWorkloadCache(t, p, workload)
	checkWorkloadCache(t, p, workloadWithoutService)
	checkWorkloadCache(t, p, workloadHostname)

	// Check Frontend Map
	checkFrontEndMapWithNetworkMode(t, workload.Addresses[0], p, workload.NetworkMode)
	checkFrontEndMapWithNetworkMode(t, workloadWithoutService.Addresses[0], p, workloadWithoutService.NetworkMode)
	checkFrontEndMapWithNetworkMode(t, workloadHostname.Addresses[0], p, workloadHostname.NetworkMode)
}

func checkWorkloadCache(t *testing.T, p *Processor, workload *workloadapi.Workload) {
	ip := workload.Addresses[0]
	address := cache.NetworkAddress{
		Network: workload.Network,
	}
	address.Address, _ = netip.AddrFromSlice(ip)
	// host network mode is not managed by kmesh
	if workload.NetworkMode == workloadapi.NetworkMode_HOST_NETWORK {
		assert.Nil(t, p.WorkloadCache.GetWorkloadByAddr(address))
	} else {
		assert.NotNil(t, p.WorkloadCache.GetWorkloadByAddr(address))
	}
	// We store pods by their uids regardless of their network mode
	assert.NotNil(t, p.WorkloadCache.GetWorkloadByUid(workload.Uid))
}

func checkServiceMap(t *testing.T, p *Processor, svcId uint32, fakeSvc *workloadapi.Service, endpointCount uint32) {
	var sv bpfcache.ServiceValue
	err := p.bpf.ServiceLookup(&bpfcache.ServiceKey{ServiceId: svcId}, &sv)
	assert.NoError(t, err)
	assert.Equal(t, endpointCount, sv.EndpointCount)
	waypointAddr := fakeSvc.GetWaypoint().GetAddress().GetAddress()
	if waypointAddr != nil {
		assert.Equal(t, test.EqualIp(sv.WaypointAddr, waypointAddr), true)
	}

	assert.Equal(t, sv.WaypointPort, nets.ConvertPortToBigEndian(fakeSvc.Waypoint.GetHboneMtlsPort()))
}

func checkBackendMap(t *testing.T, p *Processor, workloadID uint32, wl *workloadapi.Workload) {
	var bv bpfcache.BackendValue
	err := p.bpf.BackendLookup(&bpfcache.BackendKey{BackendUid: workloadID}, &bv)
	assert.NoError(t, err)
	assert.Equal(t, test.EqualIp(bv.Ip, wl.Addresses[0]), true)
	waypointAddr := wl.GetWaypoint().GetAddress().GetAddress()
	if waypointAddr != nil {
		assert.Equal(t, test.EqualIp(bv.WaypointAddr, waypointAddr), true)
	}
	assert.Equal(t, bv.WaypointPort, nets.ConvertPortToBigEndian(wl.GetWaypoint().GetHboneMtlsPort()))
}

func checkFrontEndMapWithNetworkMode(t *testing.T, ip []byte, p *Processor, networkMode workloadapi.NetworkMode) (upstreamId uint32) {
	var fk bpfcache.FrontendKey
	var fv bpfcache.FrontendValue
	nets.CopyIpByteFromSlice(&fk.Ip, ip)
	err := p.bpf.FrontendLookup(&fk, &fv)
	if networkMode != workloadapi.NetworkMode_HOST_NETWORK {
		assert.NoError(t, err)
		upstreamId = fv.UpstreamId
	} else {
		assert.Error(t, err)
	}
	return
}

func checkFrontEndMap(t *testing.T, ip []byte, p *Processor) (upstreamId uint32) {
	var fk bpfcache.FrontendKey
	var fv bpfcache.FrontendValue
	nets.CopyIpByteFromSlice(&fk.Ip, ip)
	err := p.bpf.FrontendLookup(&fk, &fv)
	assert.NoError(t, err)
	upstreamId = fv.UpstreamId
	return
}

func BenchmarkAddNewServicesWithWorkload(b *testing.B) {
	t := &testing.T{}
	config := options.BpfConfig{
		Mode:        constants.WorkloadMode,
		BpfFsPath:   "/sys/fs/bpf",
		Cgroup2Path: "/mnt/kmesh_cgroup2",
		EnableMda:   false,
	}
	cleanup, bpfLoader := test.InitBpfMap(t, config)
	b.Cleanup(cleanup)

	workloadController := NewController(bpfLoader.GetBpfKmeshWorkload())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workload := createTestWorkloadWithService(true)
		err := workloadController.Processor.handleWorkload(workload)
		assert.NoError(t, err)
	}
	workloadController.Processor.hashName.Reset()
}

func createTestWorkloadWithService(withService bool) *workloadapi.Workload {
	workload := workloadapi.Workload{
		Namespace:         "ns",
		Name:              "name",
		Addresses:         [][]byte{netip.AddrFrom4([4]byte{1, 2, 3, 4}).AsSlice()},
		Network:           "testnetwork",
		CanonicalName:     "foo",
		CanonicalRevision: "latest",
		WorkloadType:      workloadapi.WorkloadType_POD,
		WorkloadName:      "name",
		Status:            workloadapi.WorkloadStatus_HEALTHY,
		ClusterId:         "cluster0",
		Services:          map[string]*workloadapi.PortList{},
	}

	if withService == true {
		workload.Services = map[string]*workloadapi.PortList{
			"default/testsvc.default.svc.cluster.local": {
				Ports: []*workloadapi.Port{
					{
						ServicePort: 80,
						TargetPort:  8080,
					},
					{
						ServicePort: 81,
						TargetPort:  8180,
					},
					{
						ServicePort: 82,
						TargetPort:  82,
					},
				},
			},
		}
	}
	workload.Uid = "cluster0/" + rand.String(6)
	return &workload
}

func createFakeWorkload(ip string, networkMode workloadapi.NetworkMode) *workloadapi.Workload {
	workload := workloadapi.Workload{
		Namespace:         "ns",
		Name:              "name",
		Addresses:         [][]byte{netip.MustParseAddr(ip).AsSlice()},
		Network:           "testnetwork",
		CanonicalName:     "foo",
		CanonicalRevision: "latest",
		WorkloadType:      workloadapi.WorkloadType_POD,
		WorkloadName:      "name",
		Status:            workloadapi.WorkloadStatus_HEALTHY,
		ClusterId:         "cluster0",
		NetworkMode:       networkMode,
		Services: map[string]*workloadapi.PortList{
			"default/testsvc.default.svc.cluster.local": {
				Ports: []*workloadapi.Port{
					{
						ServicePort: 80,
						TargetPort:  8080,
					},
					{
						ServicePort: 81,
						TargetPort:  8180,
					},
					{
						ServicePort: 82,
						TargetPort:  82,
					},
				},
			},
		},
	}
	workload.Uid = "cluster0/" + rand.String(6)
	return &workload
}

func createFakeService(name, ip, waypoint string) *workloadapi.Service {
	return &workloadapi.Service{
		Name:      name,
		Namespace: "default",
		Hostname:  name + ".default.svc.cluster.local",
		Addresses: []*workloadapi.NetworkAddress{
			{
				Address: netip.MustParseAddr(ip).AsSlice(),
			},
		},
		Ports: []*workloadapi.Port{
			{
				ServicePort: 80,
				TargetPort:  8080,
			},
			{
				ServicePort: 81,
				TargetPort:  8180,
			},
			{
				ServicePort: 82,
				TargetPort:  82,
			},
		},
		Waypoint: &workloadapi.GatewayAddress{
			Destination: &workloadapi.GatewayAddress_Address{
				Address: &workloadapi.NetworkAddress{
					Address: netip.MustParseAddr(waypoint).AsSlice(),
				},
			},
			HboneMtlsPort: 15008,
		},
	}
}

func Test_deleteWorkloadWithRestart(t *testing.T) {
	workloadMap := bpfcache.NewFakeWorkloadMap(t)
	defer bpfcache.CleanupFakeWorkloadMap(workloadMap)

	p := newProcessor(workloadMap)

	// 1. handle workload with service, but service not handled yet
	// In this case, only frontend map and backend map should be updated.
	wl := createTestWorkloadWithService(true)
	err := p.handleWorkload(wl)
	assert.NoError(t, err)

	workloadID := checkFrontEndMap(t, wl.Addresses[0], p)
	checkBackendMap(t, p, workloadID, wl)

	epKeys := p.bpf.GetEndpointKeys(workloadID)
	assert.Equal(t, len(epKeys), 0)
	for svcName := range wl.Services {
		endpoints := p.endpointsByService[svcName]
		assert.Len(t, endpoints, 1)
		if _, ok := endpoints[wl.Uid]; ok {
			assert.True(t, ok)
		}
	}

	// Set a restart label and simulate missing data in the cache
	bpf.SetStartType(bpf.Restart)
	for key := range wl.GetServices() {
		p.ServiceCache.DeleteService(key)
	}

	p.compareWorkloadAndServiceWithHashName()
	hashNameClean(p)
}

// The hashname will be saved as a file by default.
// If it is not cleaned, it will affect other use cases.
func hashNameClean(p *Processor) {
	for str := range p.hashName.strToNum {
		if err := p.removeWorkloadFromBpfMap(str); err != nil {
			log.Errorf("RemoveWorkloadResource failed: %v", err)
		}

		if err := p.removeServiceResourceFromBpfMap(str); err != nil {
			log.Errorf("RemoveServiceResource failed: %v", err)
		}
		p.hashName.Delete(str)
	}
	p.hashName.Reset()
}
