<div align="center">
  <img
    src="https://github.com/arpansaha13/pariksha/blob/main/frontend/src/public/logo_green/transparent.png"
    width="240px"
    height="240px"
  />
  <h1>Pariksha</h1>
  <p>A comprehensive online exam platform</p>
</div>

---

## Motivation

I've created this project for learning purposes and I keep running different implementations and experiments on it 🙂.

This project uses:

- Kubernetes with Helm Charts to run all services.
- [Skaffold](https://skaffold.dev/) for development to support hot reloading with Kubernetes.
- [Hashicorp Vault](https://developer.hashicorp.com/vault) as the external secrets store.
- [External Secrets Operator](https://external-secrets.io/) to dynamically create Kubernetes Secrets from external secret store (Vault).

## Local dev

After the [project setup](https://github.com/arpansaha13/pariksha?tab=readme-ov-file#project-setup) is ready, run the below command to start the local dev server.

```bash
skaffold dev --no-prune=false --cache-artifacts=false
```

To start the kubernetes dashboard, run:

```bash
kubectl -n kubernetes-dashboard port-forward svc/kubernetes-dashboard-kong-proxy 8443:443
```

## Project Setup

> Note: The auth service, mail service, and the Vault live outside the cluster and are not included in this repository.

### Install Docker Desktop and Kubernetes (Optional)

> I am using Docker Desktop Kubernetes for local development.
> But you may use any other local Kubernetes solution that suits your environment or preferences.

- Follow this [official guide](https://docs.docker.com/desktop/setup/install/windows-install/) to install Docker Desktop.
- Follow this [official guide](https://docs.docker.com/desktop/features/kubernetes/) to enable Kubernetes in Docker Desktop.

### Setup Kubernetes Dashboard (Optional)

- Follow this [official guide](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/?trk=direct) to setup Kubernetes Dashboard on your local.

### Install Helm

- Follow this [official guide](https://helm.sh/docs/intro/install/) to install Helm.

### Install Skaffold

- Follow this [official guide](https://skaffold.dev/docs/install/) to install Skaffold.

### Install External Secrets Operator (ESO)

- Install ESO using Helm by following this [official guide](https://external-secrets.io/latest/introduction/getting-started/).

### Setup Authentication for Vault

This project uses [Kubernetes Auth](https://developer.hashicorp.com/vault/docs/auth/kubernetes) for Vault.

1. Prepare a Vault instance on your local.

2. Enable the Kubernetes Auth method.

```bash
vault auth enable -path=pariksha-kubernetes kubernetes
```

3. Run the below command at project root.

```bash
helm install pariksha-vault-auth ./vault --namespace pariksha --create-namespace
```

- This Helm Chart will create a Service Account for Token Review.
- Copy the token from the Secret bound to this Service Account.
- This will be used as the `token_reviewer_jwt` in step 6.

4. Get the `kubernetes_host` using the below command. This will be used in step 6.

```bash
kubectl cluster-info
```

> For Docker Desktop Kubernetes it will be something like "https://kubernetes.docker.internal:6443".

5. Get the `kubernetes_ca_cert` using the below command. It will be used in step 6.

```bash
kubectl get configmap kube-root-ca.crt -n kube-system -o jsonpath="{.data.ca\.crt}"
```

6. Configure the Kubernetes Auth Method uisng the values obtained from step 4, 5, and 6.

```bash
vault write auth/pariksha-kubernetes/config \
  kubernetes_host="<your local kubernetes host>" \
  token_reviewer_jwt="<your reviewer service account JWT>" \
  kubernetes_ca_cert=@<path/to/ca.crt>
```

### Create policy to access the secrets

Create the `pariksha_policy.hcl` file:

```hcl
path "pariksha/data/engine" {
  capabilities = ["read", "list"]
}

path "pariksha/data/exam" {
  capabilities = ["read", "list"]
}

path "pariksha/data/paper" {
  capabilities = ["read", "list"]
}

path "pariksha/data/question" {
  capabilities = ["read", "list"]
}
```

Write the Policy to Vault.

```bash
vault policy write pariksha_policy pariksha_policy.hcl
```

### Create the role for external-secrets

```bash
vault write auth/pariksha-kubernetes/role/external-secrets \
  bound_service_account_names="external-secrets" \
  bound_service_account_namespaces="external-secrets" \
  policies="pariksha_policy"
```

### Create KV Secrets Engine

Enable a KV (Key-Value) secrets engine at the path `pariksha`.

```bash
vault secrets enable -path=pariksha -version=2 kv
```
