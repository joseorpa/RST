package main

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer" // Helper for pointers (int32, bool, etc.)
)

// launchK6Job spins up a Pod running grafana/k6
func launchK6Job(clientset *kubernetes.Clientset, namespace string, targetURL string, virtualUsers int) error {
	
	// 1. Define the K6 Script dynamically
	// We inject the target URL using the JS string interpolation
	k6Script := fmt.Sprintf(`
		import http from 'k6/http';
		import { sleep } from 'k6';
		
		export const options = {
			vus: %d,
			duration: '30s',
		};

		export default function () {
			http.get('%s');
			sleep(1);
		}
	`, virtualUsers, targetURL)

	// 2. Generate a unique name for the job
	jobName := fmt.Sprintf("rst-attack-%d", time.Now().Unix())

	// 3. Define the Job Spec
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			// Clean up the pod 60 seconds after it finishes (Keep it clean!)
			TTLSecondsAfterFinished: pointer.Int32(60),
			BackoffLimit:            pointer.Int32(1), // Don't retry endlessly if it fails
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "rst-k6-runner"},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "k6",
							Image: "grafana/k6:latest",
							// TRICK: Echo the env var into a file, then run k6
							Command: []string{"sh", "-c", "echo \"$K6_SCRIPT\" > /script.js && k6 run /script.js"},
							Env: []corev1.EnvVar{
								{
									Name:  "K6_SCRIPT",
									Value: k6Script,
								},
							},
							Resources: corev1.ResourceRequirements{
								// Good practice to set limits
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	// 4. Send request to Kubernetes API
	fmt.Printf("Launching Job %s in namespace %s...\n", jobName, namespace)
	_, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	return err
}
