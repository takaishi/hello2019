# Hello custom controller with informer

informerを使ってイベントベースでカスタムリソースの変更に対応する。CRDについては [hello custom resource](https://github.com/takaishi/hello2018/tree/master/hello-custom-resource) や [CRDについてのメモ](https://repl.info/archives/2384/) を参照。Custom Controllerの最初の一歩については [hello-custom-controller](https://github.com/takaishi/hello2019/tree/master/hello-custom-controller) を参照

## 参考

* [Building an operator for Kubernetes with the sample-controller](https://itnext.io/building-an-operator-for-kubernetes-with-the-sample-controller-b4204be9ad56)
* 

## 準備

kubernetes v1.12.3を使う。minikubeで環境を作成

```
$ minikube start --memory=8192 --cpus=4 \
    --kubernetes-version=v1.12.3 \
    --vm-driver=hyperkit \
    --bootstrapper=kubeadm
```

- 
  go: v1.11.2

## informer用のコードを生成する

code-generatorのinformer-genを使う。

```
$ bash ~/src/k8s.io/code-generator/generate-groups.sh client,deepcopy,informer
Generating deepcopy funcs
Generating clientset for foo:v1alpha at github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/clientset
Generating informers for foo:v1alpha at github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/informers
```

## コントローラーの実装

informerを素朴に使った実装は以下の通り。リソースの追加、更新、削除に対応して任意の処理を実行できる。

```go
package main

import (
	clientset "github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/clientset/versioned"
	informers "github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	app := cli.NewApp()
	app.Flags = []cli.Flag{}

	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf(err.Error())
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Printf(err.Error())
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, time.Second*30, informers.WithNamespace("default"))
	informer := informerFactory.Samplecontroller().V1alpha().Foos()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	})

	stopCh := SetupSignalHandler()
	go informerFactory.Start(stopCh)

	log.Printf("[DEBUG] start")
	<-stopCh
	log.Printf("[DEBUG] shutdown")
	return nil
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func addFunc(obj interface{}) {
	log.Printf("[DEBUG] addFunc")
	log.Printf("[DEBUG] obj: %+v\n", obj)
}

func updateFunc(old, obj interface{}) {
	log.Printf("[DEBUG] updateFunc")
	log.Printf("[DEBUG] old: %+v\n", old)
	log.Printf("[DEBUG] obj: %+v\n", obj)
}

func deleteFunc(obj interface{}) {
	log.Printf("[DEBUG] deleteFunc")
	log.Printf("[DEBUG] obj: %+v\n", obj)
}
```

後はビルドしてデプロイするだけ。注意点として、使用するサービスアカウントがリソースをwatchできる必要があるのでRoleの更新が必要。

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: foo-reader
  namespace: default
rules:
  - apiGroups: ["samplecontroller.k8s.io"]
    verbs: ["get", "list", "watch"]
    resources: ["foos"]
```

追加時：

```
controller-main-6cbd85c8f7-8hg27 controller-main 2019/01/04 02:03:39 main.go:81: [DEBUG] addFunc
controller-main-6cbd85c8f7-8hg27 controller-main 2019/01/04 02:03:39 main.go:82: [DEBUG] obj: &{TypeMeta:{Kind: APIVersion:} ObjectMeta:{Name:foo-001 GenerateName: Namespace:default SelfLink:/apis/samplecontroller.k8s.io/v1alpha/namespaces/default/foos/foo-001 UID:f43b8abd-0fc4-11e9-95ac-263ada282756 ResourceVersion:204577 Generation:1 CreationTimestamp:2019-01-04 02:03:39 +0000 UTC DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kubectl.kubernetes.io/last-applied-configuration:{"apiVersion":"samplecontroller.k8s.io/v1alpha","kind":"Foo","metadata":{"annotations":{},"name":"foo-001","namespace":"default"},"spec":{"deploymentName":"deploy-foo-001","replicas":1}}
controller-main-6cbd85c8f7-8hg27 controller-main ] OwnerReferences:[] Initializers:nil Finalizers:[] ClusterName:} Status:{Name:} Spec:{Name:}}
```



















