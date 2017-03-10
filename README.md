# k8s-config-policy

## introduction
The purpose of this project to to audit Kubernetes resources. If k8s objects are deployed with blacklisted parameters, it will delete the object and open an issue on GitHub.

## k8s api
This project extends the Kubernetes (k8s) api by creating a `ThirdPartyResource` in which we can create our custom kind. Here is an example of how we create a `ThirdPartyResource`:

```yaml
apiVersion: extensions/v1beta1
kind: ThirdPartyResource
metadata:
  name: config-rule.frankgreco
description: "A specification for create rules on configuration files"
versions:
- name: v1
```

```yaml
---
apiVersion: k8s.io/v1
kind: ConfigPolicy
metadata:
  name: policy
  namespace: default
spec:
  apiVersion: v1
  kind: Service
  rules:
  - remove: true
    issue:
      title: Service Exposes NodePort
      body:
        issue: Due to security reasons, you are not allowed to expose a `NodePort` in this namespace. Services must be accessed via a cluster virtual ip address.
        code: "type: NodePort"
        resolution: Please remove this option and redeploy
    policy:
      template: ".spec.type"
      regex: "NodePort"

```
