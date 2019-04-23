package main

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"time"
)

func NewLeaderElector(client kubernetes.Interface) (*leaderelection.LeaderElector, error) {
	podName := os.Getenv("POD_NAME")

	callBacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			log.Printf("[INFO] started leading...")
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, _ := os.Hostname()
	source := v1.EventSource{Component: "test-leader-elector", Host: hostname}
	recorder := broadcaster.NewRecorder(scheme.Scheme, source)

	lock := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{Namespace: "default", Name: "leader-election"},
		Client: client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podName,
			EventRecorder: recorder,
		},
	}

	ttl := 30 * time.Second
	cfg := leaderelection.LeaderElectionConfig{
		Lock: &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod: ttl / 4,
		Callbacks: callBacks,
	}

	elector, err := leaderelection.NewLeaderElector(cfg)
	if err != nil {
		return nil, err
	}

	return elector, nil
}

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	elector, err := NewLeaderElector(clientset)
	ctx := context.Background()
	go elector.Run(ctx)

	os.Exit(0)
}
