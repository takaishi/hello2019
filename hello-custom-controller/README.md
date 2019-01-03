# Hello custom controller

## 準備

kubernetes v1.12.3を使う。minikubeで環境を作成

```
$ minikube start --memory=8192 --cpus=4 \
    --kubernetes-version=v1.12.3 \
    --vm-driver=hyperkit \
    --bootstrapper=kubeadm
```

## CRD

CRDについては [hello custom resource](https://github.com/takaishi/hello2018/tree/master/hello-custom-resource) や [CRDについてのメモ](https://repl.info/archives/2384/) を参照。

## Custom Controller

CRDだけではリソースを作成できるだけで何も起きない。データストアとして使うのであればCRDだけでよいが、KubernetesのdeclarativeAPIを活用する場合はコントローラーを作る必要がある。

* [Custom Resources](https://v1-12.docs.kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [KubernetesのCRD(Custom Resource Definition)とカスタムコントローラーの作成](https://qiita.com/__Attsun__/items/785008ef970ad82c679c)
* [Extending Kubernetes with Custom Resources and Operator Frameworks](https://speakerdeck.com/ianlewis/extending-kubernetes-with-custom-resources-and-operator-frameworks)
  * Kubernetesを拡張するにはデータとロジックが必要
  * データはCustomResourceDefinition
  * ロジックはController
  * Operator Framework
    * operator-sdk
    * kubebuilder
* [kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
  * `Foo` というカスタムリソースを定義する
  * このリソースは `Deployment` を定義するためのカスタムリソース
    * 名前とレプリカ数を指定できる
  * client-goライブラリを使っている
* [KubernetesのCRDまわりを整理する。](https://qiita.com/cvusk/items/773e222e0971a5391a51)



## カスタムコントローラー作っていく

カスタムコントローラーを作っていくわけだが、いきなり複雑なものを作るのは難しい。まずはカスタムリソースを読み込み、ログに出力するだけのカスタムコントローラーを作ってみる。カスタムリソース作成用のSDKやBuilderがいろいろあるようだが、これもCRD用のコードを出力するためのcode-generatorを用いて素朴に実装してみたい。

* [kubernetes/code-generator](https://github.com/kubernetes/code-generator)
* [Kubernetesを拡張しよう](https://www.ianlewis.org/jp/extending-kubernetes-ja)
* [Extending Kubernetes: Create Controllers for Core and Custom Resources](https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3)
  * コントローラのイベントフロー解説
* [KubernetesのCustom Resource Definition(CRD)とCustom Controller](https://www.sambaiz.net/article/182/)
* [Kubernetes Deep Dive: Code Generation for CustomResources](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/)
* https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go



`pkg/apis/foo`以下を作成した後、code-generatorでクライアント用コードを生成する。code-generatorはrefs/tags/kubernetes-1.12.3を使用：

```
$ env GO111MODULE=off bash ~/src/k8s.io/code-generator/generate-groups.sh all github.com/takaishi/hello2019/hello-custom-controller/pkg/client github.com/takaishi/hello2019/hello-custom-controller/pkg/apis foo:v1alpha
Generating deepcopy funcs
Generating clientset for foo:v1alpha at github.com/takaishi/hello2019/hello-custom-controller/pkg/client/clientset
Generating listers for foo:v1alpha at github.com/takaishi/hello2019/hello-custom-controller/pkg/client/listers
Generating informers for foo:v1alpha at github.com/takaishi/hello2019/hello-custom-controller/pkg/client/informers
```

これでGoからFooリソースを扱うことが出来る。カスタムコントローラーの最初の1歩として、Fooリソースを取得してログに出力してみる：

```go
package main

import (
	clientset "github.com/takaishi/hello2019/hello-custom-controller/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"fmt"
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

	for {
		foos, err := client.SamplecontrollerV1alpha().Foos("default").List(v1.ListOptions{})
		if err != nil {
			log.Printf(err.Error())
			continue
		}

		fmt.Printf("%+v\n", foos)

		time.Sleep(10 * time.Second)
	}
	return nil
}
```

コントローラー用のマニフェスト。defaultアカウントがオブジェクトを取得できるようにRBACの設定も行っている：

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: foo-reader
  namespace: default
rules:
  - apiGroups: ["samplecontroller.k8s.io"]
    verbs: ["get", "list"]
    resources: ["foos"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: foo-reader-rolebinding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: foo-reader
subjects:
  - kind: ServiceAccount
    name: default
    namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-main
spec:
  selector
    matchLabels:
      app: controller-main
  replicas: 1
  template:
    metadata:
      labels:
        app: controller-main
    spec:
      containers:
        - name: controller-main
          image: rtakaishi/sample-controller-main:latest
          imagePullPolicy: IfNotPresent
```

カスタムコントローラーのビルドからKubernetesへのデプロイまで行うためのMakefile：

```makefile
date := $(shell date +'%s')

default:
	env GOOS=linux GOARCH=amd64 go build -o controller-main main.go
	docker build . -t rtakaishi/sample-controller-main
	kubectl apply -f ./deploy-controller-main.yaml
	kubectl patch deploy controller-main -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"$(date)\"}}}}}"
```

デプロイする：

```
$ eval (minikube docker-env)
$ make
```

ログを見るとFooリソースを取得してログに出力していることが確認できた。

```
$ stern controller-main
+ controller-main-66f5cb49df-4bhq6 › controller-main
controller-main-66f5cb49df-4bhq6 controller-main &{TypeMeta:{Kind: APIVersion:} ListMeta:{SelfLink:/apis/samplecontroller.k8s.io/v1alpha/namespaces/default/foos ResourceVersion:98881 Continue:} Items:[{TypeMeta:{Kind:Foo APIVersion:samplecontroller.k8s.io/v1alpha} ObjectMeta:{Name:foo-001 GenerateName: Namespace:default SelfLink:/apis/samplecontroller.k8s.io/v1alpha/namespaces/default/foos/foo-001 UID:77506f4f-0e40-11e9-95ac-263ada282756 ResourceVersion:3649 Generation:1 CreationTimestamp:2019-01-02 03:42:45 +0000 UTC DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kubectl.kubernetes.io/last-applied-configuration:{"apiVersion":"samplecontroller.k8s.io/v1alpha","kind":"Foo","metadata":{"annotations":{},"name":"foo-001","namespace":"default"},"spec":{"deploymentName":"deploy-foo-001","replicas":1}}
controller-main-66f5cb49df-4bhq6 controller-main ] OwnerReferences:[] Initializers:nil Finalizers:[] ClusterName:} Status:{Name:} Spec:{Name:}} {TypeMeta:{Kind:Foo APIVersion:samplecontroller.k8s.io/v1alpha} ObjectMeta:{Name:foo-002 GenerateName: Namespace:default SelfLink:/apis/samplecontroller.k8s.io/v1alpha/namespaces/default/foos/foo-002 UID:775194f8-0e40-11e9-95ac-263ada282756 ResourceVersion:93582 Generation:1 CreationTimestamp:2019-01-02 03:42:45 +0000 UTC DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kubectl.kubernetes.io/last-applied-configuration:{"apiVersion":"samplecontroller.k8s.io/v1alpha","kind":"Foo","metadata":{"annotations":{},"name":"foo-002","namespace":"default"},"spec":{"deploymentName":"deploy-foo-002","hoge":"huga","replicas":1}}
controller-main-66f5cb49df-4bhq6 controller-main ] OwnerReferences:[] Initializers:nil Finalizers:[] ClusterName:} Status:{Name:} Spec:{Name:}}]}
```

