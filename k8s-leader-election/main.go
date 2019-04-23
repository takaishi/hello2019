package main

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"time"
)

func NewLeaderElector(client kubernetes.Interface, identity string) (*leaderelection.LeaderElector, error) {
	callBacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			log.Printf("[INFO] started leading...")
		},
		OnStoppedLeading: func() {
			log.Printf("[INFO] stopped leading...")
		},
		OnNewLeader: func(identity string) {
			log.Printf(fmt.Sprintf("[INFO] new leader: %s", identity))
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, _ := os.Hostname()
	source := v1.EventSource{Component: "test-leader-elector", Host: hostname}
	recorder := broadcaster.NewRecorder(scheme.Scheme, source)

	lock := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{Namespace: "default", Name: "leader-election"},
		Client:        client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      identity,
			EventRecorder: recorder,
		},
	}

	ttl := 30 * time.Second
	cfg := leaderelection.LeaderElectionConfig{
		Lock:          &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callBacks,
	}

	elector, err := leaderelection.NewLeaderElector(cfg)
	if err != nil {
		return nil, err
	}

	return elector, nil
}

func main() {
	identity := os.Args[1]
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	elector, err := NewLeaderElector(clientset, identity)
	ctx := context.Background()
	go elector.Run(ctx)

	for {
		leader := elector.GetLeader()
		isLeader := elector.IsLeader()

		fmt.Printf("[INFO] leader: %s, isLeader: %v\n", leader, isLeader)
		time.Sleep(10 * time.Second)
	}
	os.Exit(0)
}
