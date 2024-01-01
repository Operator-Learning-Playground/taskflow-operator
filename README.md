## taskflow-operator 简易型任务流控制器

### 项目思路与设计
设计背景：k8s 当中原生的 Pod 资源对象执行 container 容器时，并没有相互依赖的编排特性(ex: 容器 a 完成后 -> 再执行容器 b ...)。
本项目在此需求上，基于 k8s 的扩展功能，实现 Task 的自定义资源控制器，实现一个能顺序执行容器的 operator 应用。

```yaml
apiVersion: api.practice.com/v1alpha1
kind: Task
metadata:
  name: example-taskflow
spec:
  # 设置多个 step 步骤，会按照填入的 container 顺序执行
  steps:
    # 每项都是一个 container 对象
    - name: step1
      image: busybox:1.28
      command: [ "sh","-c" ]
      args: [ "echo step1" ]
    - name: step2
      image: busybox:1.28
      command: [ "sh","-c" ]
      args: [ "echo step2" ]
    - name: step3
      image: devopscube/kubernetes-job-demo:latest
      args: [ "100" ]
```


### 项目功能
1. 支持任务中的 container 顺序执行

### 项目部署与使用
1. 打成镜像或是使用编译二进制。
```bash
# 项目根目录执行
[root@VM-0-16-centos taskflowoperator]# pwd
/root/taskflowoperator
# 下列命令会得到一个二进制文件，服务启动时需要使用。
# 可以直接使用 docker 镜像部署
[root@VM-0-16-centos taskflowoperator]# docker build -t taskflowoperator:v1 .
Sending build context to Docker daemon  194.6kB
Step 1/15 : FROM golang:1.18.7-alpine3.15 as builder
 ---> 33c97f935029
Step 2/15 : WORKDIR /app
...
```
2. 构建 container agent 镜像
```bash
[root@VM-0-16-centos taskflowoperator]# chmod +x docker_build.sh
[root@VM-0-16-centos taskflowoperator]# ./docker_build.sh
镜像 docker.io/taskflow/agent:v1.0 不存在，开始构建...
Sending build context to Docker daemon  16.38kB
Step 1/17 : FROM golang:1.18.7-alpine3.15 as builder
 ---> 33c97f935029
Step 2/17 : RUN #mkdir /src
 ---> Using cache
 ---> abd15a740ce7
Step 3/17 : WORKDIR /app
 ---> Using cache
 ---> b25bc39a9a40
Step 4/17 : COPY go.mod go.mod
 ---> 6576edb8ae05
Step 5/17 : COPY go.sum go.sum
 ---> 3e30c14b75a4
Step 6/17 : ENV GOPROXY=https://goproxy.cn,direct
 ---> Running in d98da0a5296e
Removing intermediate container d98da0a5296e
 ---> a6263b47b54d
Step 7/17 : ENV GO111MODULE=on
 ---> Running in a6da28b7982d
Removing intermediate container a6da28b7982d
 ---> 613ebff35ca3
Step 8/17 : RUN go mod download
 ---> Running in 3b8e3931b18b
```

3. apply crd 资源
```bash
[root@VM-0-16-centos taskflowoperator]#
[root@VM-0-16-centos taskflowoperator]# kubectl apply -f deploy/task.yaml
customresourcedefinition.apiextensions.k8s.io/tasks.api.practice.com unchanged
```
4. 启动 controller 服务(需要先执行 rbac.yaml，否则服务会报错)
```bash
[root@VM-0-16-centos taskflowoperator]# kubectl apply -f deploy/task.yaml
customresourcedefinition.apiextensions.k8s.io/tasks.api.practice.com unchanged
[root@VM-0-16-centos taskflowoperator]# kubectl apply -f deploy/rbac.yaml
serviceaccount/taskflow-sa unchanged
clusterrole.rbac.authorization.k8s.io/taskflow-clusterrole unchanged
clusterrolebinding.rbac.authorization.k8s.io/taskflow-ClusterRoleBinding unchanged
[root@VM-0-16-centos taskflowoperator]# kubectl apply -f deploy/deploy.yaml
deployment.apps/taskflow-controller unchanged
```
5. 查看 operator 服务
```bash
[root@VM-0-16-centos deploy]# kubectl logs -f taskflow-controller-846ccc5bbb-w748q
I0101 05:14:59.301232       1 init_config.go:22] run in cluster!
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"controller-runtime.metrics","msg":"Metrics server is starting to listen","addr":":8080"}
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"task-flow operator","msg":"Starting server","path":"/metrics","kind":"metrics","addr":"[::]:8080"}
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"task-flow operator","msg":"Starting EventSource","controller":"task","controllerGroup":"api.practice.com","controllerKind":"Task","source":"kind source: *v1alpha1.Task"}
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"task-flow operator","msg":"Starting EventSource","controller":"task","controllerGroup":"api.practice.com","controllerKind":"Task","source":"kind source: *v1.Pod"}
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"task-flow operator","msg":"Starting Controller","controller":"task","controllerGroup":"api.practice.com","controllerKind":"Task"}
{"level":"info","ts":"2024-01-01T05:14:59Z","logger":"task-flow operator","msg":"Starting workers","controller":"task","controllerGroup":"api.practice.com","controllerKind":"Task","worker count":1}

I0101 05:15:21.106836       1 helper.go:130] step1 use normal mode.....
I0101 05:15:21.106857       1 helper.go:184] [--wait /etc/podinfo/order --waitcontent 1 --out stdout --command sh -c echo step1]
I0101 05:15:21.106875       1 helper.go:130] step2 use normal mode.....
I0101 05:15:21.106882       1 helper.go:184] [--wait /etc/podinfo/order --waitcontent 2 --out stdout --command sh -c echo step2]
I0101 05:15:21.106889       1 helper.go:130] step3 use normal mode.....

I0101 05:17:03.329371       1 image_helper.go:72] [devopscube/kubernetes-job-demo:latest] image is Image type image
I0101 05:17:07.804184       1 image_helper.go:85] Image Name: [devopscube/kubernetes-job-demo:latest], type: Image, os: [linux], Architecture: [amd64], Entrypoint: [[/script.sh]], Cmd: [[]]
I0101 05:17:07.804220       1 helper.go:184] [--wait /etc/podinfo/order --waitcontent 3 --out stdout --command /script.sh 100]
I0101 05:17:07.812416       1 task_controller.go:49] successful reconcile
I0101 05:17:07.824050       1 helper.go:241] pod status: Pending
I0101 05:17:07.824068       1 helper.go:242] annotation order:  0
I0101 05:17:07.824076       1 task_controller.go:49] successful reconcile
I0101 05:17:07.847943       1 helper.go:241] pod status: Pending
I0101 05:17:07.847963       1 helper.go:242] annotation order:  0
I0101 05:17:07.847971       1 task_controller.go:49] successful reconcile
I0101 05:17:09.756468       1 helper.go:241] pod status: Pending
I0101 05:17:09.756487       1 helper.go:242] annotation order:  0
I0101 05:17:09.756493       1 task_controller.go:49] successful reconcile
I0101 05:17:10.873731       1 task_controller.go:49] successful reconcile
I0101 05:17:10.873931       1 helper.go:241] pod status: Running
I0101 05:17:10.873942       1 helper.go:242] annotation order:  1
I0101 05:17:10.873949       1 task_controller.go:49] successful reconcile
I0101 05:17:12.141029       1 helper.go:241] pod status: Running
I0101 05:17:12.141048       1 helper.go:242] annotation order:  2
I0101 05:17:12.141057       1 task_controller.go:49] successful reconcile
I0101 05:17:12.141302       1 helper.go:241] pod status: Running
```


### RoadMap
