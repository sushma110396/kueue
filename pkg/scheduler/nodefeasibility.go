package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func (s *Scheduler) nodeCapacityFeasible(ctx context.Context, log logr.Logger, e *entry) bool {
	nodeList := &corev1.NodeList{}
	if err := s.client.List(ctx, nodeList); err != nil {
		log.Error(err, "Failed to list nodes for feasibility check")
		return true
	}

	if len(nodeList.Items) == 0 {
		return true
	}

	for _, podSet := range e.Obj.Spec.PodSets {
		podRequest := corev1.ResourceList{}

		for _, container := range podSet.Template.Spec.Containers {
			for resourceName, quantity := range container.Resources.Requests {
				if resourceName == corev1.ResourceCPU || resourceName == corev1.ResourceMemory {
					existing := podRequest[resourceName]
					existing.Add(quantity)
					podRequest[resourceName] = existing
				}
			}
		}

		fitsAnyNode := false
		for _, node := range nodeList.Items {
			nodeCPU := node.Status.Allocatable[corev1.ResourceCPU]
			nodeMemory := node.Status.Allocatable[corev1.ResourceMemory]

			if nodeCPU.Cmp(podRequest[corev1.ResourceCPU]) >= 0 &&
				nodeMemory.Cmp(podRequest[corev1.ResourceMemory]) >= 0 {
				fitsAnyNode = true
				break
			}
		}

		if !fitsAnyNode {
			log.V(2).Info("Node capacity feasibility check failed",
				"workload", klog.KObj(e.Obj),
				"podSet", podSet.Name,
				"cpuRequest", podRequest[corev1.ResourceCPU],
				"memoryRequest", podRequest[corev1.ResourceMemory],
			)
			return false
		}
	}

	return true
}
