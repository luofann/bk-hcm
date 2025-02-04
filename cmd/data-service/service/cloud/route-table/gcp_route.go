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

package routetable

import (
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud"
	tablecloud "hcm/pkg/dal/table/cloud/route-table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
	"github.com/tidwall/gjson"
)

// initGcpRouteService initialize the gcp route service.
func initGcpRouteService(svc *routeTableSvc, cap *capability.Capability) {
	h := rest.NewHandler()

	h.Path("/vendors/gcp")

	h.Add("BatchCreateGcpRoute", "POST", "/routes/batch/create", svc.BatchCreateGcpRoute)
	h.Add("ListGcpRoute", "POST", "/routes/list", svc.ListGcpRoute)
	h.Add("BatchDeleteGcpRoute", "DELETE", "/route_tables/{route_table_id}/routes/batch",
		svc.BatchDeleteGcpRoute)

	h.Load(cap.WebService)
}

// BatchCreateGcpRoute batch create route.
func (svc *routeTableSvc) BatchCreateGcpRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpRouteBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	networks := make([]string, 0)
	for _, createReq := range req.GcpRoutes {
		networks = append(networks, createReq.Network)
	}

	routeIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// generate route table
		routeTableMap, err := svc.genGcpRouteTable(cts.Kit, txn, networks)
		if err != nil {
			return nil, err
		}

		// add routes
		routes := make([]tablecloud.GcpRouteTable, 0, len(req.GcpRoutes))
		for _, createReq := range req.GcpRoutes {
			routeTable, exists := routeTableMap[createReq.Network]
			if !exists {
				return nil, errf.Newf(errf.InvalidParameter, "network(%s) route table not exists", createReq.Network)
			}

			route := tablecloud.GcpRouteTable{
				CloudID:          createReq.CloudID,
				RouteTableID:     routeTable.ID,
				VpcID:            routeTable.VpcID,
				CloudVpcID:       routeTable.CloudVpcID,
				SelfLink:         createReq.SelfLink,
				Name:             createReq.Name,
				DestRange:        createReq.DestRange,
				NextHopGateway:   createReq.NextHopGateway,
				NextHopIlb:       createReq.NextHopIlb,
				NextHopInstance:  createReq.NextHopInstance,
				NextHopIp:        createReq.NextHopIp,
				NextHopNetwork:   createReq.NextHopNetwork,
				NextHopPeering:   createReq.NextHopPeering,
				NextHopVpnTunnel: createReq.NextHopVpnTunnel,
				Priority:         createReq.Priority,
				RouteStatus:      createReq.RouteStatus,
				RouteType:        createReq.RouteType,
				Tags:             createReq.Tags,
				Memo:             createReq.Memo,
				Creator:          cts.Kit.User,
				Reviser:          cts.Kit.User,
			}

			routes = append(routes, route)
		}

		routeID, err := svc.dao.Route().Gcp().BatchCreateWithTx(cts.Kit, txn, routes)
		if err != nil {
			return nil, fmt.Errorf("create gcp route failed, err: %v", err)
		}

		return routeID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := routeIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create gcp route but return ids type %s is not string array",
			reflect.TypeOf(routeIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// genGcpRouteTable generate gcp virtual route table if it's not exists, returns the generated route table.
func (svc *routeTableSvc) genGcpRouteTable(kt *kit.Kit, txn *sqlx.Tx, networkList []string) (
	map[string]tablecloud.RouteTableTable, error) {

	// unique networks
	networkMap := make(map[string]struct{})
	networks := make([]string, 0)
	for _, network := range networkList {
		if _, exists := networkMap[network]; !exists {
			networkMap[network] = struct{}{}
			networks = append(networks, network)
		}
	}

	// list route tables by networks
	tables, err := svc.listGcpRouteTableInfo(kt, networks)
	if err != nil {
		return nil, err
	}

	tableMap := make(map[string]tablecloud.RouteTableTable)
	for _, table := range tables {
		tableMap[table.CloudVpcID] = table
		delete(networkMap, table.CloudVpcID)
	}

	// returns route tables if all exists
	if len(networkMap) == 0 {
		return tableMap, nil
	}

	// get vpc info
	notExistsNetworks := make([]string, 0)
	for network := range networkMap {
		notExistsNetworks = append(notExistsNetworks, network)
	}

	vpcs, err := svc.listGcpVpcInfo(kt, notExistsNetworks)
	if err != nil {
		return nil, err
	}

	// generate route tables by vpc
	routeTables := make([]tablecloud.RouteTableTable, 0, len(vpcs))
	for _, vpc := range vpcs {
		name := fmt.Sprintf("系统生成(%s)", converter.PtrToVal(vpc.Name))
		cloudID := fmt.Sprintf("system_generated(%s)", vpc.CloudID)

		network := gjson.Get(string(vpc.Extension), "self_link").String()

		routeTables = append(routeTables, tablecloud.RouteTableTable{
			Vendor:     enumor.Gcp,
			AccountID:  vpc.AccountID,
			CloudID:    cloudID,
			CloudVpcID: network,
			Name:       &name,
			VpcID:      vpc.ID,
			BkBizID:    constant.UnassignedBiz,
			Extension:  "{}",
			Creator:    kt.User,
			Reviser:    kt.User,
		})

		delete(networkMap, network)
	}

	if len(networkMap) != 0 {
		logs.Errorf("some networks are not exist, not exits map: %+v, rid: %s", networkMap, kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "not all networks exists")
	}

	ids, err := svc.dao.RouteTable().BatchCreateWithTx(kt, txn, routeTables)
	if err != nil {
		return nil, fmt.Errorf("create route tables failed, err: %v", err)
	}

	if len(ids) != len(routeTables) {
		return nil, errf.New(errf.RecordNotFound, "generated route table id length is invalid")
	}

	for i, table := range routeTables {
		table.ID = ids[i]
		tableMap[table.CloudVpcID] = table
	}
	return tableMap, nil
}

func (svc *routeTableSvc) listGcpRouteTableInfo(kt *kit.Kit, networks []string) ([]tablecloud.RouteTableTable, error) {
	tableOpt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "cloud_vpc_id", Op: filter.In.Factory(), Value: networks},
				filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: enumor.Gcp},
			},
		},
		Page:   &core.BasePage{Limit: uint(len(networks))},
		Fields: []string{"id", "cloud_vpc_id", "vpc_id"},
	}

	tableRes, err := svc.dao.RouteTable().List(kt, tableOpt)
	if err != nil {
		logs.Errorf("get route table by vpc(%s) failed, err: %v, rid: %s", networks, err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return tableRes.Details, nil
}

func (svc *routeTableSvc) listGcpVpcInfo(kt *kit.Kit, networks []string) ([]cloud.VpcTable, error) {
	vpcOpt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(), Value: networks},
				filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: enumor.Gcp},
			},
		},
		Page:   &core.BasePage{Limit: uint(len(networks))},
		Fields: []string{"id", "name", "account_id", "cloud_id", "extension"},
	}

	vpcRes, err := svc.dao.Vpc().List(kt, vpcOpt)
	if err != nil {
		logs.Errorf("get vpc by self link(%s) failed, err: %v, rid: %s", networks, err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return vpcRes.Details, nil
}

// ListGcpRoute list routes.
func (svc *routeTableSvc) ListGcpRoute(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.GcpRouteListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}

	if len(req.RouteTableID) != 0 {
		opt.Filter = &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: req.RouteTableID},
				req.Filter,
			},
		}
	}

	daoGcpRouteResp, err := svc.dao.Route().Gcp().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list gcp route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list gcp route failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.GcpRouteListResult{Count: daoGcpRouteResp.Count}, nil
	}

	details := make([]protocore.GcpRoute, 0, len(daoGcpRouteResp.Details))
	for _, route := range daoGcpRouteResp.Details {
		details = append(details, protocore.GcpRoute{
			ID:               route.ID,
			CloudID:          route.CloudID,
			RouteTableID:     route.RouteTableID,
			VpcID:            route.VpcID,
			CloudVpcID:       route.CloudVpcID,
			SelfLink:         route.SelfLink,
			Name:             route.Name,
			DestRange:        route.DestRange,
			NextHopGateway:   route.NextHopGateway,
			NextHopIlb:       route.NextHopIlb,
			NextHopInstance:  route.NextHopInstance,
			NextHopIp:        route.NextHopIp,
			NextHopNetwork:   route.NextHopNetwork,
			NextHopPeering:   route.NextHopPeering,
			NextHopVpnTunnel: route.NextHopVpnTunnel,
			Priority:         route.Priority,
			RouteStatus:      route.RouteStatus,
			RouteType:        route.RouteType,
			Tags:             route.Tags,
			Memo:             route.Memo,
			Revision: &core.Revision{
				Creator:   route.Creator,
				Reviser:   route.Reviser,
				CreatedAt: route.CreatedAt.String(),
				UpdatedAt: route.UpdatedAt.String(),
			},
		})
	}

	return &protocloud.GcpRouteListResult{Details: details}, nil
}

// BatchDeleteGcpRoute batch delete routes.
func (svc *routeTableSvc) BatchDeleteGcpRoute(cts *rest.Contexts) (interface{}, error) {
	tableID := cts.PathParameter("route_table_id").String()
	if tableID == "" {
		return nil, errf.New(errf.InvalidParameter, "route table id is required")
	}

	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "route_table_id", Op: filter.Equal.Factory(), Value: tableID},
				req.Filter,
			},
		},
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.Route().Gcp().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch delete list gcp route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list gcp route failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delGcpRouteIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delGcpRouteIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delGcpRouteFilter := tools.ContainersExpression("id", delGcpRouteIDs)
		if err := svc.dao.Route().Gcp().BatchDeleteWithTx(cts.Kit, txn, delGcpRouteFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete gcp route failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
