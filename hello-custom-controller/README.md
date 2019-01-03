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
Generating clientset for foo:v1alpha at github.com/takaishi/hello2018/hello-custom-resource/my-sample-controller/pkg/client/clientset
Generating listers for foo:v1alpha at github.com/takaishi/hello2018/hello-custom-resource/my-sample-controller/pkg/client/listers
Generating informers for foo:v1alpha at github.com/takaishi/hello2018/hello-custom-resource/my-sample-controller/pkg/client/informers
```
