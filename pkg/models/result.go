package models

type CurrentCfsConfig struct {
	CfsPeriodUS int32 `json:"cfs_period_us"`
	CfsQuotaUS  int32 `json:"cfs_quota_us"`
}
