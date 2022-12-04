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

package cloud

import (
	"fmt"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/protocol/base"
	"hcm/pkg/api/protocol/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"

	"hcm/pkg/dal/dao"
	tablecloud "hcm/pkg/dal/table/cloud"
)

// InitAccountService initial the account service
func InitAccountService(cap *capability.Capability) {
	svc := &accountSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	// 采用类似 iac 接口的结构简化处理, 不遵循 RESTful 风格
	h.Add("CreateAccount", "POST", "/cloud/accounts/create/", svc.Create)
	h.Add("UpdateAccounts", "POST", "/cloud/accounts/update/", svc.Update)
	h.Add("ListAccounts", "POST", "/cloud/accounts/list/", svc.List)
	h.Add("DeleteAccounts", "POST", "/cloud/accounts/delete/", svc.Delete)

	h.Load(cap.WebService)
}

type accountSvc struct {
	dao dao.Set
}

// Create create account with options
func (svc *accountSvc) Create(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.CreateAccountReq)

	if err := cts.DecodeInto(reqData); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	id, err := svc.dao.CloudAccount().Create(cts.Kit, reqData.ToModel(cts.Kit.User))
	if err != nil {
		return nil, fmt.Errorf("create cloud account failed, err: %v", err)
	}

	return &base.CreateResult{ID: id}, nil
}

// Update create account with options
func (svc *accountSvc) Update(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.UpdateAccountsReq)

	if err := cts.DecodeInto(reqData); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}
	err := svc.dao.CloudAccount().Update(cts.Kit, &reqData.FilterExpr, reqData.ToModel(cts.Kit.User))

	return nil, err
}

// List create account with options
func (svc *accountSvc) List(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.ListAccountsReq)
	if err := cts.DecodeInto(reqData); err != nil {
		return nil, err
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}
	mData, err := svc.dao.CloudAccount().List(cts.Kit, reqData.ToListOption())
	if err != nil {
		return nil, err
	}

	var details []cloud.AccountData
	for _, m := range mData {
		details = append(details, *cloud.NewAccountData(m))
	}

	return &cloud.ListAccountsResult{Details: details}, nil
}

// Delete ...
func (svc *accountSvc) Delete(cts *rest.Contexts) (interface{}, error) {
	reqData := new(cloud.DeleteAccountsReq)
	if err := cts.DecodeInto(reqData); err != nil {
		return nil, err
	}

	if err := reqData.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	err := svc.dao.CloudAccount().Delete(cts.Kit, &reqData.FilterExpr, new(tablecloud.AccountModel))

	return nil, err
}
