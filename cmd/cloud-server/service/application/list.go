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

package application

import (
	proto "hcm/pkg/api/cloud-server/application"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// List ...
func (a *applicationSvc) List(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ApplicationListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 构造过滤条件，只能查询自己的单据
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "applicant", Op: filter.Equal.Factory(), Value: cts.Kit.User},
		},
	}
	// 加上请求里过滤条件
	if req.Filter != nil && !req.Filter.IsEmpty() {
		reqFilter.Rules = append(reqFilter.Rules, req.Filter)
	}

	resp, err := a.client.DataService().Global.Application.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApplicationListReq{
			Filter: reqFilter,
			Page:   req.Page,
		},
	)
	if err != nil {
		return nil, err
	}

	// 去除content信息
	if resp != nil && len(resp.Details) > 0 {
		for _, detail := range resp.Details {
			detail.Content = ""
		}
	}

	return resp, nil
}
