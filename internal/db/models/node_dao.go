package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
)

const (
	NodeStateEnabled  = 1 // 已启用
	NodeStateDisabled = 0 // 已禁用
)

type NodeDAO dbs.DAO

func NewNodeDAO() *NodeDAO {
	return dbs.NewDAO(&NodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodes",
			Model:  new(Node),
			PkName: "id",
		},
	}).(*NodeDAO)
}

var SharedNodeDAO = NewNodeDAO()

// 启用条目
func (this *NodeDAO) EnableNode(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeDAO) DisableNode(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", NodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeDAO) FindEnabledNode(id int64) (*Node, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Node), err
}

// 根据主键查找名称
func (this *NodeDAO) FindNodeName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// 创建节点
func (this *NodeDAO) CreateNode(name string, clusterId int64) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}

	secret := rands.String(32)

	// 保存API Token
	err = SharedApiTokenDAO.CreateAPIToken(uniqueId, secret, NodeRoleNode)
	if err != nil {
		return
	}

	op := NewNodeOperator()
	op.Name = name
	op.UniqueId = uniqueId
	op.Secret = secret
	op.ClusterId = clusterId
	op.IsOn = 1
	op.State = NodeStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改节点
func (this *NodeDAO) UpdateNode(nodeId int64, name string, clusterId int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	_, err := this.Save(op)
	return err
}

// 更新节点版本
func (this *NodeDAO) UpdateNodeLatestVersion(nodeId int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.LatestVersion = dbs.SQL("latestVersion+1")
	_, err := this.Save(op)
	return err
}

// 批量更新节点版本
func (this *NodeDAO) UpdateAllNodesLatestVersionMatch(clusterId int64) error {
	nodeIds, err := this.FindAllNodeIdsMatch(clusterId)
	if err != nil {
		return err
	}
	if len(nodeIds) == 0 {
		return nil
	}
	_, err = this.Query().
		Pk(nodeIds).
		Set("latestVersion", dbs.SQL("latestVersion+1")).
		Update()
	return err
}

// 同步集群中的节点版本
func (this *NodeDAO) SyncNodeVersionsWithCluster(clusterId int64) error {
	if clusterId <= 0 {
		return errors.New("invalid cluster")
	}
	_, err := this.Query().
		Attr("clusterId", clusterId).
		Set("version", dbs.SQL("latestVersion")).
		Update()
	return err
}

// 取得有变更的集群
func (this *NodeDAO) FindChangedClusterIds() ([]int64, error) {
	ones, _, err := this.Query().
		State(NodeStateEnabled).
		Gt("latestVersion", 0).
		Where("version!=latestVersion").
		Result("DISTINCT(clusterId) AS clusterId").
		FindOnes()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, one.GetInt64("clusterId"))
	}
	return result, nil
}

// 计算所有节点数量
func (this *NodeDAO) CountAllEnabledNodes() (int64, error) {
	return this.Query().
		State(NodeStateEnabled).
		Count()
}

// 列出单页节点
func (this *NodeDAO) ListEnabledNodesMatch(offset int64, size int64, clusterId int64) (result []*Node, err error) {
	query := this.Query().
		State(NodeStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result)

	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}

	_, err = query.FindAll()
	return
}

// 根据节点ID和密钥查询节点
func (this *NodeDAO) FindEnabledNodeWithUniqueIdAndSecret(uniqueId string, secret string) (*Node, error) {
	one, err := this.Query().
		Attr("uniqueId", uniqueId).
		Attr("secret", secret).
		State(NodeStateEnabled).
		Find()

	if one != nil {
		return one.(*Node), err
	}

	return nil, err
}

// 根据节点ID获取节点
func (this *NodeDAO) FindEnabledNodeWithUniqueId(uniqueId string) (*Node, error) {
	one, err := this.Query().
		Attr("uniqueId", uniqueId).
		State(NodeStateEnabled).
		Find()

	if one != nil {
		return one.(*Node), err
	}

	return nil, err
}

// 获取节点集群ID
func (this *NodeDAO) FindNodeClusterId(nodeId int64) (int64, error) {
	col, err := this.Query().
		Pk(nodeId).
		Result("clusterId").
		FindCol(0)
	return types.Int64(col), err
}

// 匹配节点并返回节点ID
func (this *NodeDAO) FindAllNodeIdsMatch(clusterId int64) (result []int64, err error) {
	query := this.Query()
	query.State(NodeStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	query.Result("id")
	ones, _, err := query.FindOnes()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, one.GetInt64("id"))
	}
	return
}

// 计算节点数量
func (this *NodeDAO) CountAllEnabledNodesMatch(clusterId int64) (int64, error) {
	query := this.Query()
	query.State(NodeStateEnabled)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	return query.Count()
}

// 更改节点状态
func (this *NodeDAO) UpdateNodeStatus(nodeId int64, statusJSON []byte) error {
	_, err := this.Query().
		Pk(nodeId).
		Set("status", string(statusJSON)).
		Update()
	return err
}

// 设置节点安装状态
func (this *NodeDAO) UpdateNodeIsInstalled(nodeId int64, isInstalled bool) error {
	_, err := this.Query().
		Pk(nodeId).
		Set("isInstalled", isInstalled).
		Update()
	return err
}

// 生成唯一ID
func (this *NodeDAO) genUniqueId() (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query().
			Attr("uniqueId", uniqueId).
			Exist()
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		return uniqueId, nil
	}
}
