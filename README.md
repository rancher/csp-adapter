# Rancher CSP Adapter

This project is the adapter for rancher's integration with various cloud provider billing services (just does aws for now).

## Purpose

The CSP adapter tracks the number of nodes that a consumer has used and compares with the number of nodes that they are
entitled to by license purchases. If you have too many nodes, it produces a RancherUserNotification, which the UI
can display to inform the user that they need to purchase more licenses or reduce the amount of nodes being used.

The CSP adapter also produces a configmap with Cloud provider specific information (i.e. account number). This configmap
can be used by rancher to produce a supportconfig (tar which can be given to support).

## Installation

Full installation steps can be found in the rancher docs.

This chart requires:

- Rancher version 2.6.6 or higher
- Rancher is installed on an EKS cluster
- An IAM role has been configured according to the auth section of the readme and these docs
- Any private certs have been provided as described in these docs

### Certificate Setup

The adapter communicates with rancher to get accurate node counts. This communication requires that the adapter trusts rancher's certificate.

The adapter supports 2 certificate setups: standard and private.

#### Standard Certificate Setup

If rancher is using a certificate provided by a trusted Certificate Authority (i.e. letsEncrypt) no additional setup is needed.

#### Private Certificate Setup

If rancher is using a self-generated certificate or a certificate signed by a private certificate authority, you will need to provide this certificate for the adapter.

First, extract the certificate into a file called `ca-additional.pem`. If you are using the rancher generated certs option, you can use the below command:

```bash
kubectl get secret tls-rancher -n cattle-system -o jsonpath="{.data.tls\.crt}" | base64 -d  >> ca-additional.pem
```

Then, create the secret in the adapter namespace:

```bash
kubectl -n cattle-csp-adapter-system create secret generic tls-ca-additional --from-file=ca-additional.pem
```

As this certificate is rotated, you will need to replace the cert following the steps above, and then restart the adapter deployment, like below:

```bash
kubectl rollout restart deploy/rancher-csp-adapter -n cattle-csp-adapter-system
```

You can also use tools like certmanager's [trust operator](https://cert-manager.io/docs/projects/trust/) to automate this rotation. Keep in mind that this is not a supported option.

## CSP Background info 


### AWS

**License Manager**
- License manager tracks license usage through the use of entitlements
- At most, there is one "Rancher product" license in an account
- The entitlement describing how many nodes are available is the `RKE_NODE_SUPP` entitlement.
- Each `RKE_NODE_SUPP` entitles a consumer to 20 nodes (any type, includes local cluster nodes)
- Customers must manually purchase more entitlements if they use more nodes than the max allowed by `RKE_NODE_SUPP`

**Relevant API Calls**
- `ListReceivedLicenses` is used to find the licenses for the rancher support product sku
- `CheckoutLicense` is used to reserve certain entitlements for use by this rancher instance
- `ExtendLicenseConsumption` is used to extend tokens so that we can hold onto entitlements for longer than 1 hour (if not used, entitlements are automatically returned after 1 hour)
- `CheckInLicense` is used to return entitlements that are no longer being used
- `GetLicenseUsage` is used to determine how many entitlements are being used in total

**Auth**
- AWS authentication makes use of [iam roles for service accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- Because of this, you need the following setup before using the adapter:
  - An OIDC provider setup for your EKS cluster
  - An IAM role which trusts the OIDC provider
  - An IAM Role with the following permission/statement:
  ```json
  {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "license-manager:ListReceivedLicenses",
                "license-manager:CheckoutLicense",
                "license-manager:ExtendLicenseConsumption",
                "license-manager:CheckInLicense",
                "license-manager:GetLicense",
                "license-manager:GetLicenseUsage"
            ],
            "Resource": "*"
  }
  ```

## Development
`make build`

`docker build -f package/Dockerfile . -t $MY_REPO:$MY_TAG`

## Release

1. Check Kubernetes and Rancher version limits in the annotations of this repo's `charts/Chart.yaml`. Change the supported Kubernetes versions (`kube-version` range) if you have added/removed support for a version in the current range. Change the `rancher-version` range only when making a new major version of the csp-adapter.
2. Update the [rancher-csp-adapter chart](https://github.com/rancher/charts/blob/dev-v2.7/packages/rancher-csp-adapter/package.yaml):
    1. Clone this repo (not a fork).
    2. Pull the `release/v2.7` branch.
    3. Tag the branch with a new version. For example, to release version `2.2.0`, tag with `git tag v2.2.0` (make sure to include the `v`, it's important for the CI).
    4. Push the new tag to the remote: `git push --tags`. Wait for the pipeline to release the new version before moving on to the next step.
    5. In the charts repo, update the version of the adapter. See the [rancher/charts](https://github.com/rancher/charts/tree/dev-v2.7) repo for instructions on updating the [rancher-csp-adapter chart](https://github.com/rancher/charts/blob/dev-v2.7/packages/rancher-csp-adapter/package.yaml).
3. Update the value of the `CATTLE_CSP_ADAPTER_MIN_VERSION` environment variable in [Rancher's Dockerfile](https://github.com/rancher/rancher/blob/release/v2.7/package/Dockerfile).
4. Update the version compatibility matrix in the main Rancher docs.
5. Update the AWS Marketplace listing.
