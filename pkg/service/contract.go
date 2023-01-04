package service

import (
	"fmt"
	"github.com/goperate/convert/core/array"
	"github.com/hamster-shared/a-line/pkg/application"
	db2 "github.com/hamster-shared/a-line/pkg/db"
	"github.com/hamster-shared/a-line/pkg/utils"
	"github.com/hamster-shared/a-line/pkg/vo"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type ContractService struct {
	db *gorm.DB
}

func NewContractService() *ContractService {
	return &ContractService{
		db: application.GetBean[*gorm.DB]("db"),
	}
}

func (c *ContractService) SaveDeploy(entity db2.ContractDeploy) error {

	err := c.db.Transaction(func(tx *gorm.DB) error {
		tx.Save(entity)
		return nil
	})

	return err
}

func (c *ContractService) QueryContracts(projectId uint, query, version, network string, page int, size int) (vo.Page[db2.Contract], error) {
	var contracts []db2.Contract
	var afterData []db2.Contract
	sql := fmt.Sprintf("select id, project_id,workflow_id,workflow_detail_id,name,version,group_concat( DISTINCT `network` SEPARATOR ',' ) as network,build_time,abi_info,byte_code,create_time from t_contract where project_id = ? ")
	if query != "" && version != "" && network != "" {
		sql = sql + "and name like CONCAT('%',?,'%') and version = ? and network = ? group by name"
		c.db.Raw(sql, projectId, query, version, network).Scan(&contracts)
	} else if query != "" && version != "" {
		sql = sql + "and name like CONCAT('%',?,'%') and version = ? group by name"
		c.db.Raw(sql, projectId, query, version).Scan(&contracts)
	} else if query != "" && network != "" {
		sql = sql + "and name like CONCAT('%',?,'%') and network = ? group by name"
		c.db.Raw(sql, projectId, query, network).Scan(&contracts)
	} else if version != "" && network != "" {
		sql = sql + "and version = ? and network = ? group by name"
		c.db.Raw(sql, projectId, network).Scan(&contracts)
	} else if query != "" {
		sql = sql + "and name like CONCAT('%',?,'%') group by name"
		c.db.Raw(sql, projectId, query).Scan(&contracts)
	} else if network != "" {
		sql = sql + "and network = ? group by name"
		c.db.Raw(sql, projectId, network).Scan(&contracts)
	} else if version != "" {
		sql = sql + "and version = ? group by name"
		c.db.Raw(sql, projectId, version).Scan(&contracts)
	} else {
		sql = sql + "group by name"
		c.db.Raw(sql, projectId).Scan(&contracts)
	}
	if len(contracts) > 0 {
		start, end := utils.SlicePage(int64(page), int64(size), int64(len(contracts)))
		afterData = contracts[start:end]
	}
	return vo.NewPage[db2.Contract](afterData, len(contracts), page, size), nil
}

func (c *ContractService) QueryContractByWorkflow(workflowId, workflowDetailId int) ([]db2.Contract, error) {
	var contracts []db2.Contract
	res := c.db.Model(db2.Contract{}).Where("workflow_id = ? and workflow_detail_id = ?", workflowId, workflowDetailId).Find(&contracts)
	if res != nil {
		return contracts, res.Error
	}
	return contracts, nil
}

func (c *ContractService) QueryContractByVersion(projectId int, version string) ([]vo.ContractVo, error) {
	var contracts []db2.Contract
	var data []vo.ContractVo
	res := c.db.Model(db2.Contract{}).Where("project_id = ? and version = ?", projectId, version).Find(&contracts)
	if res != nil {
		return data, res.Error
	}
	if len(contracts) > 0 {
		copier.Copy(&data, &contracts)
	}
	return data, nil
}

func (c *ContractService) QueryContractDeployByVersion(projectId int, version string) (vo.ContractDeployInfoVo, error) {
	var data vo.ContractDeployInfoVo
	var contractDeployData []db2.ContractDeploy
	res := c.db.Model(db2.ContractDeploy{}).Where("project_id = ? and version = ?", projectId, version).Find(&contractDeployData)
	if res.Error != nil {
		return data, res.Error
	}
	contractInfo := make(map[string]vo.ContractInfoVo)
	if len(contractDeployData) > 0 {
		arr := array.NewObjArray(contractDeployData, "ContractId")
		res2 := arr.ToIdMapArray().(map[uint][]db2.ContractDeploy)
		for u, deploys := range res2 {
			var contractData db2.Contract
			res := c.db.Model(db2.Contract{}).Where("id = ?", u).First(&contractData)
			if res.Error == nil {
				var deployInfo []vo.DeployInfVo
				if len(deploys) > 0 {
					for _, deploy := range deploys {
						var deployData vo.DeployInfVo
						copier.Copy(&deployData, &deploy)
						deployInfo = append(deployInfo, deployData)
					}
				}
				var contractInfoVo vo.ContractInfoVo
				copier.Copy(&contractInfoVo, &contractData)
				contractInfoVo.DeployInfo = deployInfo
				contractInfo[contractData.Name] = contractInfoVo
			}
		}
	}
	data.Version = version
	data.ContractInfo = contractInfo
	return data, nil
}

func (c *ContractService) QueryVersionList(projectId int) ([]string, error) {
	var data []string
	res := c.db.Model(db2.Contract{}).Distinct("version").Select("version").Where("project_id = ?", projectId).Find(&data)
	if res.Error != nil {
		return data, res.Error
	}
	return data, nil
}

func (c *ContractService) QueryContractNameList(projectId int) ([]string, error) {
	var data []string
	res := c.db.Model(db2.Contract{}).Distinct("name").Select("name").Where("project_id = ?", projectId).Find(&data)
	if res.Error != nil {
		return data, res.Error
	}
	return data, nil
}

func (c *ContractService) QueryNetworkList(projectId int) ([]string, error) {
	var data []string
	res := c.db.Model(db2.Contract{}).Distinct("network").Select("network").Where("project_id = ?", projectId).Find(&data)
	if res.Error != nil {
		return data, res.Error
	}
	return data, nil
}
