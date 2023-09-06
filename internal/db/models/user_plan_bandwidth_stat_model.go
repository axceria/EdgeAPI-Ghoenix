package models

import "github.com/iwind/TeaGo/dbs"

const (
	UserPlanBandwidthStatField_Id                  dbs.FieldName = "id"                  // ID
	UserPlanBandwidthStatField_UserId              dbs.FieldName = "userId"              // 用户ID
	UserPlanBandwidthStatField_UserPlanId          dbs.FieldName = "userPlanId"          // 用户套餐ID
	UserPlanBandwidthStatField_Day                 dbs.FieldName = "day"                 // 日期YYYYMMDD
	UserPlanBandwidthStatField_TimeAt              dbs.FieldName = "timeAt"              // 时间点HHII
	UserPlanBandwidthStatField_Bytes               dbs.FieldName = "bytes"               // 带宽
	UserPlanBandwidthStatField_RegionId            dbs.FieldName = "regionId"            // 区域ID
	UserPlanBandwidthStatField_TotalBytes          dbs.FieldName = "totalBytes"          // 总流量
	UserPlanBandwidthStatField_AvgBytes            dbs.FieldName = "avgBytes"            // 平均流量
	UserPlanBandwidthStatField_CachedBytes         dbs.FieldName = "cachedBytes"         // 缓存的流量
	UserPlanBandwidthStatField_AttackBytes         dbs.FieldName = "attackBytes"         // 攻击流量
	UserPlanBandwidthStatField_CountRequests       dbs.FieldName = "countRequests"       // 请求数
	UserPlanBandwidthStatField_CountCachedRequests dbs.FieldName = "countCachedRequests" // 缓存的请求数
	UserPlanBandwidthStatField_CountAttackRequests dbs.FieldName = "countAttackRequests" // 攻击请求数
)

// UserPlanBandwidthStat 用户套餐带宽峰值
type UserPlanBandwidthStat struct {
	Id                  uint64 `field:"id"`                  // ID
	UserId              uint64 `field:"userId"`              // 用户ID
	UserPlanId          uint64 `field:"userPlanId"`          // 用户套餐ID
	Day                 string `field:"day"`                 // 日期YYYYMMDD
	TimeAt              string `field:"timeAt"`              // 时间点HHII
	Bytes               uint64 `field:"bytes"`               // 带宽
	RegionId            uint32 `field:"regionId"`            // 区域ID
	TotalBytes          uint64 `field:"totalBytes"`          // 总流量
	AvgBytes            uint64 `field:"avgBytes"`            // 平均流量
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存的流量
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存的请求数
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击请求数
}

type UserPlanBandwidthStatOperator struct {
	Id                  any // ID
	UserId              any // 用户ID
	UserPlanId          any // 用户套餐ID
	Day                 any // 日期YYYYMMDD
	TimeAt              any // 时间点HHII
	Bytes               any // 带宽
	RegionId            any // 区域ID
	TotalBytes          any // 总流量
	AvgBytes            any // 平均流量
	CachedBytes         any // 缓存的流量
	AttackBytes         any // 攻击流量
	CountRequests       any // 请求数
	CountCachedRequests any // 缓存的请求数
	CountAttackRequests any // 攻击请求数
}

func NewUserPlanBandwidthStatOperator() *UserPlanBandwidthStatOperator {
	return &UserPlanBandwidthStatOperator{}
}
