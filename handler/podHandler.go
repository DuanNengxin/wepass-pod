package handler

import (
	"context"
	"gitee.com/duannengxin/wepass-pod/common"
	"gitee.com/duannengxin/wepass-pod/domain/model"
	"gitee.com/duannengxin/wepass-pod/domain/service"
	pod "gitee.com/duannengxin/wepass-pod/proto"
	"go.uber.org/zap"
)

type PodHandler struct {
	PodDataService service.IPodDataService
}

func (p PodHandler) AddPod(ctx context.Context, info *pod.PodInfo, response *pod.Response) error {
	podModel := &model.Pod{}
	err := common.SwapTo(info, podModel)
	if err != nil {
		zap.S().Errorf("AddPod swap error %s", err.Error())
		response.Msg = err.Error()
		return err
	}

	err = p.PodDataService.CreateToK8s(info)
	if err != nil {
		zap.S().Errorf("AddPod create to k8s error %s", err.Error())
		response.Msg = err.Error()
		return err
	}
	podID, err := p.PodDataService.AddPod(podModel)
	if err != nil {
		zap.S().Errorf("AddPod PodDataService error %s", err.Error())
		response.Msg = err.Error()
		return err
	}
	zap.S().Infof("AddPod success pod id %d", podID)
	return nil
}

func (p PodHandler) DeletePod(ctx context.Context, id *pod.PodID, response *pod.Response) error {
	podModel, err := p.PodDataService.FindPodByID(id.GetId())
	if err != nil {
		zap.S().Errorf("find pod %d error %s", id.GetId(), err.Error())
		response.Msg = err.Error()
		return err
	}
	if err := p.PodDataService.DeleteToK8s(podModel); err != nil {
		zap.S().Errorf("delete pod %d %s error %s", id.GetId(), podModel.PodName, err.Error())
		response.Msg = err.Error()
		return err
	}
	if err := p.PodDataService.DeletePod(id.GetId()); err != nil {
		zap.S().Errorf("db delete pod %d error %s", id.GetId(), err.Error())
		response.Msg = err.Error()
	}
	return nil
}

func (p PodHandler) FindPodByID(ctx context.Context, id *pod.PodID, info *pod.PodInfo) error {
	podModel, err := p.PodDataService.FindPodByID(id.GetId())
	if err != nil {
		zap.S().Errorf("FindPodByID pod id %d error %s", id.GetId(), err.Error())
		return err
	}
	err = common.SwapTo(podModel, info)
	if err != nil {
		zap.S().Errorf("FindPodByID swap to error %s", err.Error())
		return err
	}
	return nil
}

func (p PodHandler) UpdatePod(ctx context.Context, info *pod.PodInfo, response *pod.Response) error {
	err := p.PodDataService.UpdateToK8s(info)
	if err != nil {
		zap.S().Errorf("UpdatePod create to k8s error %s", err.Error())
		response.Msg = err.Error()
		return err
	}

	podModel := &model.Pod{}
	err = common.SwapTo(info, podModel)
	if err != nil {
		zap.S().Errorf("UpdatePod swap error %s", err.Error())
		response.Msg = err.Error()
		return err
	}

	err = p.PodDataService.UpdatePod(podModel)
	if err != nil {
		zap.S().Errorf("UpdatePod PodDataService error %s", err.Error())
		response.Msg = err.Error()
		return err
	}
	zap.S().Infof("UpdatePod success pod id %d", info.Id)
	return nil
}

func (p PodHandler) FindPodAll(ctx context.Context, all *pod.FindAll, infos *pod.PodInfos) error {
	podModels, err := p.PodDataService.FindAll()
	if err != nil {
		zap.S().Errorf("FindPodByID pods error %s", err.Error())
		return err
	}

	for _, podModel := range podModels {
		podInfo := &pod.PodInfo{}
		err := common.SwapTo(podModel, podInfo)
		if err != nil {
			zap.S().Errorf("FindPodAll swap error %s", err.Error())
			return err
		}
		infos.PodInfos = append(infos.PodInfos, podInfo)
	}
	return nil
}
