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





















