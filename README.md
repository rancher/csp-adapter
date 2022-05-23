# Rancher CSP Adapter

This project is the adapter for rancher's integration with various cloud provider billing services (just does aws for now).

## Purpose

The CSP adapter tracks the number of nodes that a consumer has used and compares with the number of nodes that they are
entitled to by license purchases. If you have too many nodes, it produces a RancherUserNotification, which the UI
can display to inform the user that they need to purchase more licenses or reduce the amount of nodes being used.

The CSP adapter also produces a configmap with Cloud provider specific information (i.e. account number). This configmap
can be used by rancher to produce a supportconfig (tar which can be given to support).

## Installation
NOTE: The CSP adapter requires rancher to be installed before use (rancher provides node metrics and other important features).

TODO

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