# ultimate-service

## Up and running
*Make sure you have installed docker and kind on your local machine*

Up and running local kubernetes cluster
```shell
make kind up
```

Build image
```shell
make docker-build
```

Load images into the cluster
```shell
make kind-load
```

Running the deployment
```shell
make sales-apply
```

Restart the deployment (it rebuilds the image)
```shell
make sales-restart
```

Pods status
```shell
make sales-status
```

Pod logs
```shell
make sales-log
```
