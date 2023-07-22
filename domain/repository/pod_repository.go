package repository

import (
	"github.com/DuanNengxin/wepass-pod/domain/model"
	"gorm.io/gorm"
)

type IPodRepository interface {
	InitTable() error
	FindPodByID(id int64) (*model.Pod, error)
	CreatePod(pod *model.Pod) (int64, error)
	DeletePod(id int64) error
	UpdatePod(pod *model.Pod) error
	FindAll() ([]*model.Pod, error)
}

type PodRepository struct {
	mysqlDb *gorm.DB
}

func NewPodRepository(db *gorm.DB) IPodRepository {
	return &PodRepository{mysqlDb: db}
}

func (p PodRepository) InitTable() error {
	return p.mysqlDb.Migrator().CreateTable(&model.PodPort{}, &model.PodEnv{}, &model.Pod{})
}

func (p PodRepository) FindPodByID(id int64) (*model.Pod, error) {
	var pod model.Pod
	err := p.mysqlDb.Preload("PodEnv").Preload("PodPort").Find(&pod, id)
	if err != nil {
		return nil, err.Error
	}
	return &pod, nil
}

func (p PodRepository) CreatePod(pod *model.Pod) (int64, error) {
	return pod.ID, p.mysqlDb.Create(pod).Error
}

func (p PodRepository) DeletePod(id int64) error {
	tx := p.mysqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Callback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}
	if err := p.mysqlDb.Delete(&model.Pod{}, id).Error; err != nil {
		tx.Callback()
		return err
	}
	if err := p.mysqlDb.Where("pod_id = ?", id).Delete(&model.PodPort{}).Error; err != nil {
		tx.Callback()
		return err
	}
	if err := p.mysqlDb.Where("pod_id = ?", id).Delete(&model.PodEnv{}).Error; err != nil {
		tx.Callback()
		return err
	}

	return tx.Commit().Error
}

func (p PodRepository) UpdatePod(pod *model.Pod) error {
	return p.mysqlDb.Updates(pod).Error
}

func (p PodRepository) FindAll() ([]*model.Pod, error) {
	var pods []*model.Pod
	if err := p.mysqlDb.Find(&pods).Error; err != nil {
		return nil, err
	}
	return pods, nil
}
