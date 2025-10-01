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

	// svc := &corev1.Service{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "ingress-lb",
	// 		Namespace: "kube-system",
	// 	},
	// 	Spec: corev1.ServiceSpec{
	// 		Type:                  corev1.ServiceTypeLoadBalancer,
	// 		ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
	// 		InternalTrafficPolicy: (*corev1.ServiceInternalTrafficPolicyType)(func() *corev1.ServiceInternalTrafficPolicyType {
	// 			t := corev1.ServiceInternalTrafficPolicyCluster
	// 			return &t
	// 		}()),
	// 		Ports: []corev1.ServicePort{
	// 			{
	// 				Name:       "https",
	// 				Port:       443,
	// 				Protocol:   corev1.ProtocolTCP,
	// 				TargetPort: intstrFromInt(443),
	// 			},
	// 		},
	// 		Selector: map[string]string{
	// 			"app": "ingress-nginx",
	// 		},
	// 		SessionAffinity: corev1.ServiceAffinityNone,
	// 	},
	// }

	// created, err := clientset.CoreV1().Services("kube-system").Create(context.Background(), svc, metav1.CreateOptions{})
	// if err != nil {
	// 	log.Fatalf("Failed to create service: %v", err)
	// }

}
