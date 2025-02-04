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

package hcservice

import (
	"hcm/pkg/criteria/validator"
)

// TCloudAccountCheckReq ...
type TCloudAccountCheckReq struct {
	CloudMainAccountID string `json:"cloud_main_account_id" validate:"required"`
	CloudSubAccountID  string `json:"cloud_sub_account_id" validate:"required"`
	CloudSecretID      string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey     string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *TCloudAccountCheckReq) Validate() error {
	// TODO: 是否还需要添加其他规则校验呢？
	return validator.Validate.Struct(r)
}

// AwsAccountCheckReq ...
type AwsAccountCheckReq struct {
	CloudAccountID   string `json:"cloud_account_id" validate:"required"`
	CloudIamUsername string `json:"cloud_iam_username" validate:"required"`
	CloudSecretID    string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey   string `json:"cloud_secret_key" validate:"required"`
}

// Validate ...
func (r *AwsAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiAccountCheckReq ...
type HuaWeiAccountCheckReq struct {
	CloudMainAccountName string `json:"cloud_main_account_name" validate:"required"`
	CloudSubAccountID    string `json:"cloud_sub_account_id" validate:"required"`
	CloudSubAccountName  string `json:"cloud_sub_account_name" validate:"required"`
	CloudSecretID        string `json:"cloud_secret_id" validate:"required"`
	CloudSecretKey       string `json:"cloud_secret_key" validate:"required"`
	CloudIamUserID       string `json:"cloud_iam_user_id" validate:"required"`
	CloudIamUsername     string `json:"cloud_iam_username" validate:"required"`
}

// Validate ...
func (r *HuaWeiAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpAccountCheckReq ...
type GcpAccountCheckReq struct {
	CloudProjectID        string `json:"cloud_project_id" validate:"required"`
	CloudServiceSecretKey string `json:"cloud_service_secret_key" validate:"required"`
}

// Validate ...
func (r *GcpAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureAccountCheckReq ...
type AzureAccountCheckReq struct {
	CloudTenantID        string `json:"cloud_tenant_id" validate:"required"`
	CloudSubscriptionID  string `json:"cloud_subscription_id" validate:"required"`
	CloudApplicationID   string `json:"cloud_application_id" validate:"required"`
	CloudClientSecretKey string `json:"cloud_client_secret_key" validate:"required"`
}

// Validate ...
func (r *AzureAccountCheckReq) Validate() error {
	return validator.Validate.Struct(r)
}
