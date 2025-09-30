package kubernetes

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Helper) CreateSvc(svc *corev1.Service) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	return h.SvcClient.Create(ctx, svc, metav1.CreateOptions{})
}
