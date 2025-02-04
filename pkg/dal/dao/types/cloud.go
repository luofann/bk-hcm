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

package types

import "hcm/pkg/criteria/enumor"

// CloudResourceBasicInfo define cloud resource basic info.
type CloudResourceBasicInfo struct {
	ID        string        `json:"id" db:"id"`
	Vendor    enumor.Vendor `json:"vendor" db:"vendor"`
	AccountID string        `json:"account_id" db:"account_id"`
	BkBizID   int64         `json:"bk_biz_id" db:"bk_biz_id"`
	// these fields are basic info for some resource, needs to be specified explicitly.
	Region        string `json:"region" db:"region"`
	RecycleStatus string `json:"recycle_status" db:"recycle_status"`
}

// CommonBasicInfoFields defines common cloud resource basic info fields.
var CommonBasicInfoFields = []string{"id", "vendor", "account_id", "bk_biz_id"}
