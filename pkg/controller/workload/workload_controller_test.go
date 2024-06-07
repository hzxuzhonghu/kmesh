/*
 * Copyright 2024 The Kmesh Authors.
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
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	cluster_v2 "kmesh.net/kmesh/api/v2/cluster"
	core_v2 "kmesh.net/kmesh/api/v2/core"
	"kmesh.net/kmesh/pkg/auth"
	"kmesh.net/kmesh/pkg/controller/xdstest"
)

func TestWorkloadStreamCreateAndSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	workloadStream := Controller{
		Processor: nil,
	}

	// create a fake grpc service client
	mockDiscovery := xdstest.NewMockServer(t)
	conn, err := grpc.Dial("buffcon",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return mockDiscovery.Listener.Dial()
		}))
	if err != nil {
		t.Errorf("grpc connection client create failed, %s", err)
	}
	defer conn.Close()
	client := discoveryv3.NewAggregatedDiscoveryServiceClient(conn)
	stream, streamErr := client.DeltaAggregatedResources(ctx)
	if streamErr != nil {
		t.Errorf("create stream failed, %s", streamErr)
	}
	workloadStream.Stream = stream

	patches1 := gomonkey.NewPatches()
	patches2 := gomonkey.NewPatches()
	tests := []struct {
		name       string
		beforeFunc func()
		afterFunc  func()
		wantErr    bool
	}{
		{
			name: "test1: send request failed, should return error",
			beforeFunc: func() {
				patches1.ApplyMethod(reflect.TypeOf(client), "DeltaAggregatedResources",
					func(_ discoveryv3.AggregatedDiscoveryServiceClient, ctx context.Context, opts ...grpc.CallOption) (discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient, error) {
						return stream, nil
					})
				patches2.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Send",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient, req *discoveryv3.DeltaDiscoveryRequest) error {
						return errors.New("timeout")
					})
			},
			afterFunc: func() {
				patches1.Reset()
				patches2.Reset()
			},
			wantErr: true,
		},
		{
			name: "test2: no mock, send request successful, should return nil",
			beforeFunc: func() {
				cluster := &cluster_v2.Cluster{
					ApiStatus:      core_v2.ApiStatus_UPDATE,
					Name:           "ut-cluster1",
					ConnectTimeout: uint32(30),
					LbPolicy:       cluster_v2.Cluster_RANDOM,
				}
				anyCluster, err := anypb.New(cluster)
				if err != nil {
					t.Fatal("failed to create anyCluster:", err)
				}
				mockDiscovery.DeltaResponses <- &discoveryv3.DeltaDiscoveryResponse{
					Resources: []*discoveryv3.Resource{
						{
							Resource: anyCluster,
						},
					},
				}
			},
			afterFunc: func() {},
			wantErr:   false,
		},
		{
			name: "test3: fail to create workloadStream, should return error",
			beforeFunc: func() {
				patches1.ApplyMethod(reflect.TypeOf(client), "DeltaAggregatedResources",
					func(_ discoveryv3.AggregatedDiscoveryServiceClient, ctx context.Context, opts ...grpc.CallOption) (discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient, error) {
						return nil, errors.New("fail to create adsstream")
					})
			},
			afterFunc: func() {
				patches1.Reset()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc()
			err := workloadStream.WorkloadStreamCreateAndSend(client, context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("worklaodStream.WorklaodStreamCreateAndSend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.afterFunc()
		})
	}
}

func TestAdsStream_AdsStreamProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	workloadStream := Controller{
		Processor: &Processor{
			ack: &discoveryv3.DeltaDiscoveryRequest{},
		},
	}

	// create a fake grpc service client
	mockDiscovery := xdstest.NewMockServer(t)
	conn, err := grpc.Dial("buffcon",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return mockDiscovery.Listener.Dial()
		}))
	if err != nil {
		t.Errorf("grpc connection client create failed, %s", err)
	}
	defer conn.Close()
	client := discoveryv3.NewAggregatedDiscoveryServiceClient(conn)
	stream, streamErr := client.DeltaAggregatedResources(ctx)
	if streamErr != nil {
		t.Errorf("create stream failed, %s", streamErr)
	}
	workloadStream.Stream = stream

	patches1 := gomonkey.NewPatches()
	patches2 := gomonkey.NewPatches()
	tests := []struct {
		name       string
		beforeFunc func()
		afterFunc  func()
		wantErr    bool
	}{
		{
			name: "test1: stream Revc failed, should return error",
			beforeFunc: func() {
				patches1.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Recv",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient) (*discoveryv3.DeltaDiscoveryResponse, error) {
						return nil, errors.New("failed to recv message")
					})
			},
			afterFunc: func() {
				patches1.Reset()
			},
			wantErr: true,
		},
		{
			name: "test2: stream Send failed, should return error",
			beforeFunc: func() {
				patches1.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Recv",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient) (*discoveryv3.DeltaDiscoveryResponse, error) {
						// create resource of rsq
						cluster := &config_cluster_v3.Cluster{
							Name: "ut-cluster",
						}
						anyCluster, _ := anypb.New(cluster)
						return &discoveryv3.DeltaDiscoveryResponse{
							TypeUrl: AddressType,
							Nonce:   "222",
							Resources: []*discoveryv3.Resource{
								{
									Resource: anyCluster,
								},
							},
						}, nil
					})
				patches2.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Send",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient) error {
						return errors.New("failed to send message")
					})
			},
			afterFunc: func() {
				patches1.Reset()
				patches2.Reset()
			},
			wantErr: true,
		},
		{
			name: "test3: handle success, should return nil",
			beforeFunc: func() {
				patches1.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Recv",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient) (*discoveryv3.DeltaDiscoveryResponse, error) {
						// create resource of rsq
						cluster := &config_cluster_v3.Cluster{
							Name: "ut-cluster",
						}
						anyCluster, _ := anypb.New(cluster)
						return &discoveryv3.DeltaDiscoveryResponse{
							TypeUrl: AddressType,
							Nonce:   "222",
							Resources: []*discoveryv3.Resource{
								{
									Resource: anyCluster,
								},
							},
						}, nil
					})
				patches2.ApplyMethod(reflect.TypeOf(workloadStream.Stream), "Send",
					func(_ discoveryv3.AggregatedDiscoveryService_DeltaAggregatedResourcesClient) error {
						return nil
					})
			},
			afterFunc: func() {
				patches1.Reset()
				patches2.Reset()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beforeFunc()
			rbac := auth.Rbac{}
			err := workloadStream.HandleWorkloadStream(&rbac)

			if (err != nil) != tt.wantErr {
				t.Errorf("workloadStream.WorkloadStreamProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.afterFunc()
		})
	}
}
