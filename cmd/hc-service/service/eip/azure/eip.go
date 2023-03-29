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

package azure

import (
	"strings"

	"hcm/cmd/hc-service/logics/sync/cvm"
	synceip "hcm/cmd/hc-service/logics/sync/eip"
	syncnetworkinterface "hcm/cmd/hc-service/logics/sync/network-interface"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/eip/datasvc"
	"hcm/pkg/adaptor/azure"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/eip"
	apicore "hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	hcservice "hcm/pkg/api/hc-service"
	proto "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// EipSvc ...
type EipSvc struct {
	Adaptor *cloudclient.CloudAdaptorClient
	DataCli *dataservice.Client
}

// DeleteEip ...
func (svc *EipSvc) DeleteEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.EipDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt, err := svc.makeEipDeleteOption(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	err = client.DeleteEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.EipManager{DataCli: svc.DataCli}
	return nil, manager.Delete(cts.Kit, []string{req.EipID})
}

// AssociateEip ...
func (svc *EipSvc) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureEipAssociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt, err := svc.makeEipAssociateOption(cts.Kit, req, client)
	if err != nil {
		return nil, err
	}

	err = client.AssociateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	if err = manager.Create(cts.Kit); err != nil {
		return nil, err
	}

	_, err = synceip.SyncAzureEip(
		cts.Kit,
		&synceip.SyncAzureEipOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{opt.CloudEipID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncAzureEip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	_, err = cvm.SyncAzureCvmWithRelResource(
		cts.Kit,
		svc.Adaptor,
		svc.DataCli,
		&cvm.SyncAzureCvmOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{cvmData.CloudID},
		},
	)
	if err != nil {
		logs.Errorf("SyncAzureCvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	networkInterface, err := svc.DataCli.Azure.NetworkInterface.Get(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		req.NetworkInterfaceID,
	)
	if err != nil {
		return nil, err
	}

	_, err = syncnetworkinterface.AzureNetworkInterfaceSync(
		cts.Kit,
		&hcservice.AzureNetworkInterfaceSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{networkInterface.CloudID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("AzureNetworkInterfaceSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateEip ...
func (svc *EipSvc) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureEipDisassociateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt, err := svc.makeEipDisassociateOption(cts.Kit, req, client)
	if err != nil {
		return nil, err
	}

	err = client.DisassociateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	manager := datasvc.EipCvmRelManager{CvmID: req.CvmID, EipID: req.EipID, DataCli: svc.DataCli}
	if err = manager.Delete(cts.Kit); err != nil {
		return nil, err
	}

	_, err = synceip.SyncAzureEip(
		cts.Kit,
		&synceip.SyncAzureEipOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{opt.CloudEipID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncAzureEip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	cvmData, err := svc.DataCli.Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), req.CvmID)
	if err != nil {
		return nil, err
	}

	_, err = cvm.SyncAzureCvm(
		cts.Kit,
		svc.Adaptor,
		svc.DataCli,
		&cvm.SyncAzureCvmOption{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{cvmData.CloudID},
		},
	)
	if err != nil {
		logs.Errorf("SyncAzureCvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	networkInterface, err := svc.DataCli.Azure.NetworkInterface.Get(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		req.NetworkInterfaceID,
	)
	if err != nil {
		return nil, err
	}

	_, err = syncnetworkinterface.AzureNetworkInterfaceSync(
		cts.Kit,
		&hcservice.AzureNetworkInterfaceSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          []string{networkInterface.CloudID},
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("AzureNetworkInterfaceSync failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// CreateEip ...
func (svc *EipSvc) CreateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AzureEipCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.Adaptor.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt, err := svc.makeEipCreateOption(req)
	if err != nil {
		return nil, err
	}

	eipPtr, err := client.CreateEip(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	cloudIDs := []string{strings.ToLower(converter.PtrToVal(eipPtr))}

	_, err = synceip.SyncAzureEip(
		cts.Kit,
		&synceip.SyncAzureEipOption{
			AccountID:         req.AccountID,
			CloudIDs:          cloudIDs,
			ResourceGroupName: req.ResourceGroupName,
		},
		svc.Adaptor,
		svc.DataCli,
	)
	if err != nil {
		logs.Errorf("SyncAzureEip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	resp, err := svc.DataCli.Global.ListEip(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.EipListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: cloudIDs,
				}, &filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: string(enumor.Azure),
				},
			},
		}, Page: &apicore.BasePage{Limit: uint(len(cloudIDs))}, Fields: []string{"id"}},
	)

	eipIDs := make([]string, len(cloudIDs))
	for idx, eipData := range resp.Details {
		eipIDs[idx] = eipData.ID
	}

	return &apicore.BatchCreateResult{IDs: eipIDs}, nil
}

func (svc *EipSvc) makeEipDeleteOption(
	kt *kit.Kit,
	req *proto.EipDeleteReq,
) (*eip.AzureEipDeleteOption, error) {
	eipData, err := svc.DataCli.Azure.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}
	return &eip.AzureEipDeleteOption{
		ResourceGroupName: eipData.Extension.ResourceGroupName,
		EipName:           *eipData.Name,
	}, nil
}

func (svc *EipSvc) makeEipAssociateOption(
	kt *kit.Kit,
	req *proto.AzureEipAssociateReq,
	cli *azure.Azure,
) (*eip.AzureEipAssociateOption, error) {
	dataCli := svc.DataCli.Azure

	eipData, err := dataCli.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	networkInterface, err := dataCli.NetworkInterface.Get(kt.Ctx, kt.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}
	networkInterfaces, err := cli.ListRawNetworkInterfaceByIDs(
		kt,
		&core.AzureListByIDOption{CloudIDs: []string{networkInterface.CloudID}},
	)
	if err != nil {
		return nil, err
	}

	return &eip.AzureEipAssociateOption{
		ResourceGroupName: eipData.Extension.ResourceGroupName,
		CloudEipID:        eipData.CloudID,
		NetworkInterface:  networkInterfaces[0],
	}, nil
}

func (svc *EipSvc) makeEipDisassociateOption(
	kt *kit.Kit,
	req *proto.AzureEipDisassociateReq,
	cli *azure.Azure,
) (*eip.AzureEipDisassociateOption, error) {
	dataCli := svc.DataCli.Azure

	eipData, err := dataCli.RetrieveEip(kt.Ctx, kt.Header(), req.EipID)
	if err != nil {
		return nil, err
	}

	networkInterface, err := dataCli.NetworkInterface.Get(kt.Ctx, kt.Header(), req.NetworkInterfaceID)
	if err != nil {
		return nil, err
	}
	networkInterfaces, err := cli.ListRawNetworkInterfaceByIDs(
		kt,
		&core.AzureListByIDOption{CloudIDs: []string{networkInterface.CloudID}},
	)
	if err != nil {
		return nil, err
	}

	return &eip.AzureEipDisassociateOption{
		ResourceGroupName: eipData.Extension.ResourceGroupName,
		CloudEipID:        eipData.CloudID,
		NetworkInterface:  networkInterfaces[0],
	}, nil
}

func (svc *EipSvc) makeEipCreateOption(req *proto.AzureEipCreateReq) (*eip.AzureEipCreateOption, error) {
	return &eip.AzureEipCreateOption{
		ResourceGroupName:    req.ResourceGroupName,
		EipName:              req.EipName,
		Region:               req.Region,
		Zone:                 req.Zone,
		SKUName:              req.SKUName,
		SKUTier:              req.SKUTier,
		AllocationMethod:     req.AllocationMethod,
		IPVersion:            req.IPVersion,
		IdleTimeoutInMinutes: req.IdleTimeoutInMinutes,
	}, nil
}
