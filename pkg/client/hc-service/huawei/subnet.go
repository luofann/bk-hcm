/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package huawei

import (
	"context"
	"net/http"

	"hcm/pkg/api/core"
	proto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// SubnetClient is hc service huawei subnet api client.
type SubnetClient struct {
	client rest.ClientInterface
}

// NewSubnetClient create a new subnet api client.
func NewSubnetClient(client rest.ClientInterface) *SubnetClient {
	return &SubnetClient{
		client: client,
	}
}

// Create subnet.
func (s *SubnetClient) Create(ctx context.Context, h http.Header,
	req *proto.SubnetCreateReq[proto.HuaWeiSubnetCreateExt]) (*core.CreateResult, error) {

	resp := new(core.CreateResp)

	err := s.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/create").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// Update subnet.
func (s *SubnetClient) Update(ctx context.Context, h http.Header, id string, req *proto.SubnetUpdateReq) error {
	resp := new(rest.BaseResp)

	err := s.client.Patch().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/%s", id).
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// Delete subnet.
func (s *SubnetClient) Delete(ctx context.Context, h http.Header, id string) error {
	resp := new(rest.BaseResp)

	err := s.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResourcef("/subnets/%s", id).
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// SyncSubnet sync huawei subnet.
func (s *SubnetClient) SyncSubnet(ctx context.Context, h http.Header, req *sync.HuaWeiSubnetSyncReq) error {
	resp := new(rest.BaseResp)

	err := s.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef("/subnets/sync").
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// CountIP count huawei subnet available ips.
func (s *SubnetClient) CountIP(ctx context.Context, h http.Header, id string) (*proto.AvailIPResult, error) {
	resp := new(proto.GetAvailIPResp)

	err := s.client.Post().
		WithContext(ctx).
		Body(nil).
		SubResourcef("/subnets/%s/ips/count", id).
		WithHeaders(h).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}
