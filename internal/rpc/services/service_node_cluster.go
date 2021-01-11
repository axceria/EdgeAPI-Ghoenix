package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"strconv"
)

type NodeClusterService struct {
	BaseService
}

// 创建集群
func (this *NodeClusterService) CreateNodeCluster(ctx context.Context, req *pb.CreateNodeClusterRequest) (*pb.CreateNodeClusterResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	systemServices := map[string]maps.Map{}
	if len(req.SystemServicesJSON) > 0 {
		err = json.Unmarshal(req.SystemServicesJSON, &systemServices)
		if err != nil {
			return nil, err
		}
	}

	var clusterId int64
	err = this.RunTx(func(tx *dbs.Tx) error {
		clusterId, err = models.SharedNodeClusterDAO.CreateCluster(tx, adminId, req.Name, req.GrantId, req.InstallDir, req.DnsDomainId, req.DnsName, req.HttpCachePolicyId, req.HttpFirewallPolicyId, systemServices)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeClusterResponse{NodeClusterId: clusterId}, nil
}

// 修改集群
func (this *NodeClusterService) UpdateNodeCluster(ctx context.Context, req *pb.UpdateNodeClusterRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateCluster(tx, req.NodeClusterId, req.Name, req.GrantId, req.InstallDir)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 禁用集群
func (this *NodeClusterService) DeleteNodeCluster(ctx context.Context, req *pb.DeleteNodeClusterRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.DisableNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找单个集群
func (this *NodeClusterService) FindEnabledNodeCluster(ctx context.Context, req *pb.FindEnabledNodeClusterRequest) (*pb.FindEnabledNodeClusterResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 检查用户是否有权限
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil {
		return &pb.FindEnabledNodeClusterResponse{}, nil
	}

	return &pb.FindEnabledNodeClusterResponse{NodeCluster: &pb.NodeCluster{
		Id:                   int64(cluster.Id),
		Name:                 cluster.Name,
		CreatedAt:            int64(cluster.CreatedAt),
		InstallDir:           cluster.InstallDir,
		GrantId:              int64(cluster.GrantId),
		UniqueId:             cluster.UniqueId,
		Secret:               cluster.Secret,
		HttpCachePolicyId:    int64(cluster.CachePolicyId),
		HttpFirewallPolicyId: int64(cluster.HttpFirewallPolicyId),
		DnsName:              cluster.DnsName,
		DnsDomainId:          int64(cluster.DnsDomainId),
	}}, nil
}

// 查找集群的API节点信息
func (this *NodeClusterService) FindAPINodesWithNodeCluster(ctx context.Context, req *pb.FindAPINodesWithNodeClusterRequest) (*pb.FindAPINodesWithNodeClusterResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("can not find cluster with id '" + strconv.FormatInt(req.NodeClusterId, 10) + "'")
	}

	result := &pb.FindAPINodesWithNodeClusterResponse{}
	result.UseAllAPINodes = cluster.UseAllAPINodes == 1

	apiNodeIds := []int64{}
	if len(cluster.ApiNodes) > 0 && cluster.ApiNodes != "null" {
		err = json.Unmarshal([]byte(cluster.ApiNodes), &apiNodeIds)
		if err != nil {
			return nil, err
		}
		if len(apiNodeIds) > 0 {
			apiNodes := []*pb.APINode{}
			for _, apiNodeId := range apiNodeIds {
				apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, apiNodeId)
				if err != nil {
					return nil, err
				}
				apiNodeAddrs, err := apiNode.DecodeAccessAddrStrings()
				if err != nil {
					return nil, err
				}
				apiNodes = append(apiNodes, &pb.APINode{
					Id:            int64(apiNode.Id),
					IsOn:          apiNode.IsOn == 1,
					NodeClusterId: int64(apiNode.ClusterId),
					Name:          apiNode.Name,
					Description:   apiNode.Description,
					AccessAddrs:   apiNodeAddrs,
				})
			}
			result.ApiNodes = apiNodes
		}
	}

	return result, nil
}

// 查找所有可用的集群
func (this *NodeClusterService) FindAllEnabledNodeClusters(ctx context.Context, req *pb.FindAllEnabledNodeClustersRequest) (*pb.FindAllEnabledNodeClustersResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
			UniqueId:  cluster.UniqueId,
			Secret:    cluster.Secret,
		})
	}

	return &pb.FindAllEnabledNodeClustersResponse{
		NodeClusters: result,
	}, nil
}

// 查找所有变更的集群
func (this *NodeClusterService) FindAllChangedNodeClusters(ctx context.Context, req *pb.FindAllChangedNodeClustersRequest) (*pb.FindAllChangedNodeClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusterIds, err := models.SharedNodeDAO.FindChangedClusterIds(tx)
	if err != nil {
		return nil, err
	}
	if len(clusterIds) == 0 {
		return &pb.FindAllChangedNodeClustersResponse{
			NodeClusters: []*pb.NodeCluster{},
		}, nil
	}
	result := []*pb.NodeCluster{}
	for _, clusterId := range clusterIds {
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, clusterId)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
			UniqueId:  cluster.UniqueId,
			Secret:    cluster.Secret,
		})
	}
	return &pb.FindAllChangedNodeClustersResponse{NodeClusters: result}, nil
}

// 计算所有集群数量
func (this *NodeClusterService) CountAllEnabledNodeClusters(ctx context.Context, req *pb.CountAllEnabledNodeClustersRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClusters(tx, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 列出单页集群
func (this *NodeClusterService) ListEnabledNodeClusters(ctx context.Context, req *pb.ListEnabledNodeClustersRequest) (*pb.ListEnabledNodeClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.ListEnabledClusters(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:          int64(cluster.Id),
			Name:        cluster.Name,
			CreatedAt:   int64(cluster.CreatedAt),
			GrantId:     int64(cluster.GrantId),
			InstallDir:  cluster.InstallDir,
			UniqueId:    cluster.UniqueId,
			Secret:      cluster.Secret,
			DnsName:     cluster.DnsName,
			DnsDomainId: int64(cluster.DnsDomainId),
		})
	}

	return &pb.ListEnabledNodeClustersResponse{NodeClusters: result}, nil
}

// 查找集群的健康检查配置
func (this *NodeClusterService) FindNodeClusterHealthCheckConfig(ctx context.Context, req *pb.FindNodeClusterHealthCheckConfigRequest) (*pb.FindNodeClusterHealthCheckConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	config, err := models.SharedNodeClusterDAO.FindClusterHealthCheckConfig(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindNodeClusterHealthCheckConfigResponse{HealthCheckJSON: configJSON}, nil
}

// 修改集群健康检查设置
func (this *NodeClusterService) UpdateNodeClusterHealthCheck(ctx context.Context, req *pb.UpdateNodeClusterHealthCheckRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterHealthCheck(tx, req.NodeClusterId, req.HealthCheckJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 执行健康检查
func (this *NodeClusterService) ExecuteNodeClusterHealthCheck(ctx context.Context, req *pb.ExecuteNodeClusterHealthCheckRequest) (*pb.ExecuteNodeClusterHealthCheckResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	executor := tasks.NewHealthCheckExecutor(req.NodeClusterId)
	results, err := executor.Run()
	if err != nil {
		return nil, err
	}
	pbResults := []*pb.ExecuteNodeClusterHealthCheckResponse_Result{}
	for _, result := range results {
		pbResults = append(pbResults, &pb.ExecuteNodeClusterHealthCheckResponse_Result{
			Node: &pb.Node{
				Id:   int64(result.Node.Id),
				Name: result.Node.Name,
			},
			NodeAddr: result.NodeAddr,
			IsOk:     result.IsOk,
			Error:    result.Error,
			CostMs:   types.Float32(result.CostMs),
		})
	}
	return &pb.ExecuteNodeClusterHealthCheckResponse{Results: pbResults}, nil
}

// 计算使用某个认证的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithGrantId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithGrantIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithGrantId(tx, req.GrantId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 查找使用某个认证的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithGrantId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithGrantIdRequest) (*pb.FindAllEnabledNodeClustersWithGrantIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithGrantId(tx, req.GrantId)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
			UniqueId:  cluster.UniqueId,
			Secret:    cluster.Secret,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithGrantIdResponse{NodeClusters: result}, nil
}

// 查找集群的DNS配置
func (this *NodeClusterService) FindEnabledNodeClusterDNS(ctx context.Context, req *pb.FindEnabledNodeClusterDNSRequest) (*pb.FindEnabledNodeClusterDNSResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	dnsInfo, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	if dnsInfo == nil {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:     "",
			Domain:   nil,
			Provider: nil,
		}, nil
	}

	dnsConfig, err := dnsInfo.DecodeDNSConfig()
	if err != nil {
		return nil, err
	}

	if dnsInfo.DnsDomainId == 0 {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:            dnsInfo.DnsName,
			Domain:          nil,
			Provider:        nil,
			NodesAutoSync:   dnsConfig.NodesAutoSync,
			ServersAutoSync: dnsConfig.ServersAutoSync,
		}, nil
	}

	domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, int64(dnsInfo.DnsDomainId))
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.FindEnabledNodeClusterDNSResponse{
			Name:     dnsInfo.DnsName,
			Domain:   nil,
			Provider: nil,
		}, nil
	}
	pbDomain := &pb.DNSDomain{
		Id:   int64(domain.Id),
		Name: domain.Name,
		IsOn: domain.IsOn == 1,
	}

	provider, err := models.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(domain.ProviderId))
	if err != nil {
		return nil, err
	}

	var pbProvider *pb.DNSProvider = nil
	if provider != nil {
		pbProvider = &pb.DNSProvider{
			Id:       int64(provider.Id),
			Name:     provider.Name,
			Type:     provider.Type,
			TypeName: dnsclients.FindProviderTypeName(provider.Type),
		}
	}

	return &pb.FindEnabledNodeClusterDNSResponse{
		Name:            dnsInfo.DnsName,
		Domain:          pbDomain,
		Provider:        pbProvider,
		NodesAutoSync:   dnsConfig.NodesAutoSync,
		ServersAutoSync: dnsConfig.ServersAutoSync,
	}, nil
}

// 计算使用某个DNS服务商的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithDNSProviderId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithDNSProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 计算使用某个DNS域名的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithDNSDomainId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithDNSDomainIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClustersWithDNSDomainId(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 查找使用某个域名的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithDNSDomainId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithDNSDomainIdRequest) (*pb.FindAllEnabledNodeClustersWithDNSDomainIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledClustersWithDNSDomainId(tx, req.DnsDomainId)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:          int64(cluster.Id),
			Name:        cluster.Name,
			DnsName:     cluster.DnsName,
			DnsDomainId: int64(cluster.DnsDomainId),
		})
	}
	return &pb.FindAllEnabledNodeClustersWithDNSDomainIdResponse{NodeClusters: result}, nil
}

// 检查集群域名是否已经被使用
func (this *NodeClusterService) CheckNodeClusterDNSName(ctx context.Context, req *pb.CheckNodeClusterDNSNameRequest) (*pb.CheckNodeClusterDNSNameResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	exists, err := models.SharedNodeClusterDAO.ExistClusterDNSName(tx, req.DnsName, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.CheckNodeClusterDNSNameResponse{IsUsed: exists}, nil
}

// 修改集群的域名设置
func (this *NodeClusterService) UpdateNodeClusterDNS(ctx context.Context, req *pb.UpdateNodeClusterDNSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterDNS(tx, req.NodeClusterId, req.DnsName, req.DnsDomainId, req.NodesAutoSync, req.ServersAutoSync)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 检查集群的DNS是否有变化
func (this *NodeClusterService) CheckNodeClusterDNSChanges(ctx context.Context, req *pb.CheckNodeClusterDNSChangesRequest) (*pb.CheckNodeClusterDNSChangesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	cluster, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil || len(cluster.DnsName) == 0 || cluster.DnsDomainId <= 0 {
		return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: false}, nil
	}

	domainId := int64(cluster.DnsDomainId)
	domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: false}, nil
	}
	records, err := domain.DecodeRecords()
	if err != nil {
		return nil, err
	}

	service := &DNSDomainService{}
	changes, _, _, _, _, _, _, err := service.findClusterDNSChanges(cluster, records, domain.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CheckNodeClusterDNSChangesResponse{IsChanged: len(changes) > 0}, nil
}

// 查找集群的TOA配置
func (this *NodeClusterService) FindEnabledNodeClusterTOA(ctx context.Context, req *pb.FindEnabledNodeClusterTOARequest) (*pb.FindEnabledNodeClusterTOAResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	tx := this.NullTx()

	config, err := models.SharedNodeClusterDAO.FindClusterTOAConfig(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledNodeClusterTOAResponse{ToaJSON: configJSON}, nil
}

// 修改集群的TOA设置
func (this *NodeClusterService) UpdateNodeClusterTOA(ctx context.Context, req *pb.UpdateNodeClusterTOARequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateClusterTOA(tx, req.NodeClusterId, req.ToaJSON)
	if err != nil {
		return nil, err
	}

	// 增加节点版本号
	err = models.SharedNodeDAO.IncreaseAllNodesLatestVersionMatch(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算使用某个缓存策略的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithHTTPCachePolicyId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithHTTPCachePolicyIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledNodeClustersWithHTTPCachePolicyId(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 查找使用缓存策略的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithHTTPCachePolicyId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdRequest) (*pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.NodeCluster{}
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledNodeClustersWithHTTPCachePolicyId(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithHTTPCachePolicyIdResponse{
		NodeClusters: result,
	}, nil
}

// 计算使用某个WAF策略的集群数量
func (this *NodeClusterService) CountAllEnabledNodeClustersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.CountAllEnabledNodeClustersWithHTTPFirewallPolicyIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedNodeClusterDAO.CountAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 查找使用WAF策略的所有集群
func (this *NodeClusterService) FindAllEnabledNodeClustersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdRequest) (*pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.NodeCluster{}
	clusters, err := models.SharedNodeClusterDAO.FindAllEnabledNodeClustersWithHTTPFirewallPolicyId(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
		})
	}
	return &pb.FindAllEnabledNodeClustersWithHTTPFirewallPolicyIdResponse{
		NodeClusters: result,
	}, nil
}

// 修改集群的缓存策略
func (this *NodeClusterService) UpdateNodeClusterHTTPCachePolicyId(ctx context.Context, req *pb.UpdateNodeClusterHTTPCachePolicyIdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateNodeClusterHTTPCachePolicyId(tx, req.NodeClusterId, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}

	// 增加节点版本号
	err = models.SharedNodeDAO.IncreaseAllNodesLatestVersionMatch(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改集群的WAF策略
func (this *NodeClusterService) UpdateNodeClusterHTTPFirewallPolicyId(ctx context.Context, req *pb.UpdateNodeClusterHTTPFirewallPolicyIdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeClusterDAO.UpdateNodeClusterHTTPFirewallPolicyId(tx, req.NodeClusterId, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}

	// 增加节点版本号
	err = models.SharedNodeDAO.IncreaseAllNodesLatestVersionMatch(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改集群的系统服务设置
func (this *NodeClusterService) UpdateNodeClusterSystemService(ctx context.Context, req *pb.UpdateNodeClusterSystemServiceRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()
	err = models.SharedNodeClusterDAO.UpdateNodeClusterSystemService(tx, req.NodeClusterId, req.Type, params)
	if err != nil {
		return nil, err
	}

	// 增加节点版本号
	err = models.SharedNodeDAO.IncreaseAllNodesLatestVersionMatch(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找集群的系统服务设置
func (this *NodeClusterService) FindNodeClusterSystemService(ctx context.Context, req *pb.FindNodeClusterSystemServiceRequest) (*pb.FindNodeClusterSystemServiceResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	params, err := models.SharedNodeClusterDAO.FindNodeClusterSystemServiceParams(tx, req.NodeClusterId, req.Type)
	if err != nil {
		return nil, err
	}
	if params == nil {
		params = maps.Map{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &pb.FindNodeClusterSystemServiceResponse{ParamsJSON: paramsJSON}, nil
}
