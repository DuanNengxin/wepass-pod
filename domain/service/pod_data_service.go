package service

import (
	"context"
	"errors"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"pod/domain/model"
	"pod/domain/repository"
	pod "pod/proto"
	"strconv"
)

type IPodDataService interface {
	AddPod(pod *model.Pod) (int64, error)
	UpdatePod(pod *model.Pod) error
	DeletePod(int64) error
	FindPodByID(id int64) (*model.Pod, error)
	FindAll() ([]*model.Pod, error)
	CreateToK8s(info *pod.PodInfo) error
	UpdateToK8s(info *pod.PodInfo) error
	DeleteToK8s(*model.Pod) error
}

type PodDataService struct {
	PodRepository repository.IPodRepository
	K8sClientSet  *kubernetes.Clientset
	deployment    *appsv1.Deployment
}

func NewPodDataService(podRepository repository.IPodRepository, clientSet *kubernetes.Clientset) IPodDataService {
	return &PodDataService{
		PodRepository: podRepository,
		K8sClientSet:  clientSet,
		deployment:    &appsv1.Deployment{},
	}
}

func (p PodDataService) AddPod(pod *model.Pod) (int64, error) {
	return p.PodRepository.CreatePod(pod)
}

func (p PodDataService) UpdatePod(pod *model.Pod) error {
	return p.PodRepository.UpdatePod(pod)
}

func (p PodDataService) DeletePod(id int64) error {
	return p.PodRepository.DeletePod(id)
}

func (p PodDataService) FindPodByID(id int64) (*model.Pod, error) {
	return p.PodRepository.FindPodByID(id)
}

func (p PodDataService) FindAll() ([]*model.Pod, error) {
	return p.PodRepository.FindAll()
}

func (p PodDataService) CreateToK8s(info *pod.PodInfo) error {
	p.SetDeployment(info)
	if _, err := p.K8sClientSet.AppsV1().Deployments(info.PodNamespace).Create(context.TODO(), p.deployment, metav1.CreateOptions{}); err != nil {
		return err
	} else {
		//可以写自己的业务逻辑
		zap.S().Error("Pod " + info.PodName + "已经存在")
		return errors.New("Pod " + info.PodName + " 已经存在")
	}
}

func (p *PodDataService) SetDeployment(info *pod.PodInfo) {
	deployment := appsv1.Deployment{}
	deployment.TypeMeta = metav1.TypeMeta{
		Kind:       "deployment",
		APIVersion: "v1",
	}
	deployment.ObjectMeta = metav1.ObjectMeta{
		Name:      info.PodName,
		Namespace: info.PodNamespace,
		Labels: map[string]string{
			"app-name": info.PodName,
		},
		//Annotations: nil,
	}
	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: &info.PodReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app-name": info.PodName},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app_name": info.PodName},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            info.PodName,
						Image:           info.PodImage,
						Ports:           p.getContainerPort(info),
						Env:             p.getContainerEnvVar(info.PodEnv),
						Resources:       p.getResource(info),
						ImagePullPolicy: p.getImagePullPolicy(info.PodPullPolicy),
					},
				},
				RestartPolicy: p.getRestartPolicy(info.PodRestart),
			},
		},
		Strategy:                appsv1.DeploymentStrategy{},
		MinReadySeconds:         0,
		RevisionHistoryLimit:    nil,
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
	}
	p.deployment = &deployment
	return
}

func (p *PodDataService) getContainerPort(info *pod.PodInfo) []corev1.ContainerPort {
	var ports []corev1.ContainerPort
	for _, podPort := range info.PodPort {
		containerPort := corev1.ContainerPort{
			Name:          "port-" + strconv.Itoa(int(podPort.ContainerPort)),
			ContainerPort: podPort.ContainerPort,
			Protocol:      p.getProtocol(podPort.Protocol),
		}
		ports = append(ports, containerPort)
	}
	return ports
}

func (p *PodDataService) getProtocol(protocol string) corev1.Protocol {
	switch protocol {
	case "TCP":
		return "TCP"
	case "UDP":
		return "UDP"
	case "SCTP":
		return "SCTP"
	default:
		return "TCP"
	}
}

func (p *PodDataService) getContainerEnvVar(podEnv []*pod.PodEnv) (envs []corev1.EnvVar) {
	for _, env := range podEnv {
		envs = append(envs, corev1.EnvVar{
			Name:  env.EnvKey,
			Value: env.EnvValue,
		})
	}
	return
}

func (p *PodDataService) getResource(podInfo *pod.PodInfo) (resourceRequirement corev1.ResourceRequirements) {
	resourceRequirement.Limits = corev1.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax), 'f', 6, 64)),
	}
	resourceRequirement.Requests = corev1.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMin), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMin), 'f', 6, 64)),
	}
	return
}

func (p *PodDataService) getImagePullPolicy(policy string) corev1.PullPolicy {
	switch policy {
	case "Always":
		return corev1.PullAlways
	case "Never":
		return corev1.PullNever
	case "IfNotPresent":
		return corev1.PullIfNotPresent
	default:
		return corev1.PullAlways
	}
}

func (p *PodDataService) getRestartPolicy(policy string) corev1.RestartPolicy {
	switch policy {
	case "Always":
		return corev1.RestartPolicyAlways
	case "OnFailure":
		return corev1.RestartPolicyOnFailure
	case "Never":
		return corev1.RestartPolicyNever
	default:
		return corev1.RestartPolicyAlways
	}
}

func (p PodDataService) UpdateToK8s(info *pod.PodInfo) error {
	p.SetDeployment(info)
	if _, err := p.K8sClientSet.AppsV1().Deployments(info.PodNamespace).Update(context.TODO(), p.deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (p PodDataService) DeleteToK8s(pod *model.Pod) error {
	if err := p.K8sClientSet.AppsV1().Deployments(pod.PodNamespace).Delete(context.TODO(), pod.PodName, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
