package member

import (
	"smlcloudplatform/internal/member/models"
	"smlcloudplatform/pkg/microservice"
)

type IMemberPGRepository interface {
	Count(shopID string, guid string) (int, error)
	Create(doc models.MemberIndex) error
	Delete(shopID string, guid string) error
	FindByGuid(shopID string, guid string) (models.MemberIndex, error)
}

type MemberPGRepository struct {
	pst microservice.IPersister
}

func NewMemberPGRepository(pst microservice.IPersister) MemberPGRepository {
	return MemberPGRepository{
		pst: pst,
	}
}
func (repo MemberPGRepository) Count(shopID string, guid string) (int, error) {
	count, err := repo.pst.Count(models.MemberIndex{}, " shop_id = ? AND guid_fixed = ?", shopID, guid)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (repo MemberPGRepository) Create(member models.MemberIndex) error {
	err := repo.pst.Create(member)
	if err != nil {
		return err
	}
	return nil
}

func (repo MemberPGRepository) Delete(shopID string, guid string) error {
	tableName := models.MemberIndex{}.TableName()
	err := repo.pst.Exec("DELETE FROM "+tableName+" WHERE shop_id = ? AND guid_fixed = ?", shopID, guid)
	if err != nil {
		return err
	}
	return nil
}

func (repo MemberPGRepository) FindByGuid(shopID string, guid string) (models.MemberIndex, error) {
	inv := models.MemberIndex{}
	_, err := repo.pst.Where(&inv, "  shop_id = ? AND guid_fixed = ?", shopID, guid)
	if err != nil {
		return inv, err
	}
	return inv, nil
}
