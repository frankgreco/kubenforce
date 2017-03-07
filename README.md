# k8s-audit

## introduction
The purpose of this project to to audit deployments. If k8s objects are deployed with blacklisted parameters, it will delete the object and open an issue on GitHub.

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
Here is how we might use this to create our custom kind. The purpose of our custom kind is to specify a series of rules we want to be applied to new configuration files that are applied to a k8s cluster.

In this example, we specify that no pod in our cluster can expose a host network. By itself this wouldn't be feasible as a lot of the core components of k8s use the host network (e.g. kube-proxy). Hence, as part of the spec we can specify what namespaces we want to exclude from this rule. Optionally, you can specify what namespaces you want to apply the rule to.

```yaml
apiVersion: "frankgreco/v1"
kind: ConfigRule
metadata:
  name: my-new-rule
spec:
  rules:
  - kind: "pod"
    alert: "github"
    namespace:
      except:
      - "default"
      - "kube-system"
      include:
      - "application"
    issue:
      title: "Pod enables hostNetwork"
      body:
        issue: "Due to security reasons, Pods cannot utilize the `hostNetwork` option in this Namespace."
        code: "hostNetwork: true"
        resolution: "Please disable this options and redeploy"
      template:
        location: pod.Spec.HostNetwork
        accept:
        - true
```
