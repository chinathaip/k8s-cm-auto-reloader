package reloader

import (
	corev1 "k8s.io/api/core/v1"
)

func useConfigMap(containers []corev1.Container, ref string) bool {
	for _, c := range containers {
		for _, e := range c.Env {
			if e.ValueFrom.ConfigMapKeyRef.Name == ref {
				return true
			}
		}
	}
	return false
}
