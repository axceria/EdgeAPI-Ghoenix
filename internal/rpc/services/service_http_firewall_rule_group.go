package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// WAF规则分组相关服务
type HTTPFirewallRuleGroupService struct {
	BaseService
}

// 设置是否启用分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupIsOnRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, req.FirewallRuleGroupId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 创建分组
func (this *HTTPFirewallRuleGroupService) CreateHTTPFirewallRuleGroup(ctx context.Context, req *pb.CreateHTTPFirewallRuleGroupRequest) (*pb.CreateHTTPFirewallRuleGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroup(tx, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPFirewallRuleGroupResponse{FirewallRuleGroupId: groupId}, nil
}

// 修改分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroup(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroup(tx, req.FirewallRuleGroupId, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 获取分组配置
func (this *HTTPFirewallRuleGroupService) FindEnabledHTTPFirewallRuleGroupConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleGroupConfigRequest) (*pb.FindEnabledHTTPFirewallRuleGroupConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupConfig, err := models.SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, req.FirewallRuleGroupId)
	if err != nil {
		return nil, err
	}
	if groupConfig == nil {
		return &pb.FindEnabledHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: nil}, nil
	}
	groupConfigJSON, err := json.Marshal(groupConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: groupConfigJSON}, nil
}

// 获取分组信息
func (this *HTTPFirewallRuleGroupService) FindEnabledHTTPFirewallRuleGroup(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleGroupRequest) (*pb.FindEnabledHTTPFirewallRuleGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	group, err := models.SharedHTTPFirewallRuleGroupDAO.FindEnabledHTTPFirewallRuleGroup(tx, req.FirewallRuleGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledHTTPFirewallRuleGroupResponse{
			FirewallRuleGroup: nil,
		}, nil
	}

	return &pb.FindEnabledHTTPFirewallRuleGroupResponse{
		FirewallRuleGroup: &pb.HTTPFirewallRuleGroup{
			Id:          int64(group.Id),
			Name:        group.Name,
			IsOn:        group.IsOn == 1,
			Description: group.Description,
			Code:        group.Code,
		},
	}, nil
}

// 修改分组的规则集
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupSets(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupSetsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(tx, req.GetFirewallRuleGroupId(), req.FirewallRuleSetsJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
