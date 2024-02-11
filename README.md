# expose-k8s-operator
The operator will create a deployment, a service, and an ingress resource to expose it on a desired port when an **Expose** resource is created by an administrator.

## Description
The operator’s job is to create a deployment, a service, and an ingress resource to expose it on a desired port when a custom resource is created by an administrator. The idea is the same as the controller built by [Mr. Vivek Singh](https://www.youtube.com/playlist?list=PLh4KH3LtJvRQ43JAwwjvTnsVOMp0WKnJO), where he wrote a logic to create a service and ingress resources when a deployment resource is created without using CRD's. Also this operator is bootstrapped with Kubebuilder.

## Getting Started

### Prerequisites
- go version v1.20.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### 'Expose' Custom Resource

```
apiVersion: api.core.expose-k8s-operator.io/v1alpha1
kind: Expose
metadata:
  labels:
    app.kubernetes.io/name: expose
    app.kubernetes.io/instance: expose-sample
    app.kubernetes.io/part-of: expose-k8s-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: expose-k8s-operator
  name: expose-sample
spec:
  name: cr-dsi
  deployment:
    - name: nginx-deployment
      replicas: 1
      component: nginx
      containers:
        - name: nginx
          image: nginx
  service:
    - name: nginx-service
      port: 80
  ingress:
    - name: nginx-ingress
      path: /nginx

```

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/expose-k8s-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=roopeshsn/expose-k8s-operator:1.0.0-alpha
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
kubectl apply -f config/samples/api_v1alpha1_expose.yaml
```

**To delete a custom resource**
```sh
kubectl delete Expose expose-sample -n default
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Contributing
Thanks for taking the time to contribute to `expose-k8s-operator`! You can work on existing issues or propose to work on a new feature.

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023 roopeshsn.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

