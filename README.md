# Billing Service

## Requirements

- Information on support config and generating this
- Metrics exposed from Rancher
- AWS Product and licensing info
- How the UI will interact with billing container
- QA and local testing strategies need to be more well defined
- More information needed around entitlements to nail down license manager logic
- IAM permissions for licenses specific to Rancher for LicenseManager API calls
- Criteria for which nodes to debit, i.e. all nodes or only active ones
- Investigate containerization of Rancher support script
  - audit against suse support config, figure out what's different / required
  - Robert Schweikert can help answer questions here
  - separate feature / separate project / separate container

## Overview

The billing service will be a stateless service with a few main application components:
- collector
- manager
- api

### Collector

The collector's job is to scrape the Rancher server's `/metrics` endpoint to gather node usage data.
The Rancher metrics endpoint is authenticated, so a service account will need to be used. 
The collector will calculate the node counts from the metrics for several pre-defined buckets.

#### Buckets

1. Hosted Nodes
- this will include counts from `<metric>{provider="gke|aks|eks", type="downstream"}`

2. Management Nodes
- this will include counts from `<metric>{provider="all", type="management"}`

3. RKE Nodes
- this will include counts from `<metric>{provider="rke|rke2|rke.windows|k3s", type="downstream"}`

4. Longhorn Nodes
- this will include counts from `<metric>{provider="longhorn", type="all"}`

### Manager

The manager runs a reconciliation loop that ensures the appropriate licenses are checked out from the AWS License Manager.

Upon startup, the manager will do the following:
- List licenses for each product SKU we care about (predefined / hard-coded)
- Checkout 1 rancher start pack from the starter pack licenses (should only be 1 of these)
- Obtain node metrics from collector
- Checkout `ceiling((gke_nodes+eks_nodes+aks_nodes)/10)` hosted node 10 packs from the hosted licenses
- Checkout `ceiling((rke_nodes+k3s_nodes+rke2_nodes+rke_windows_nodes)/10)` rke node 10 packs from the rke licenses
- Checkout `ceiling(longhorn_nodes/10)` longhorn node 10 packs from the longhorn licenses
- Start reconciliation loop, passing license info for all the checked out licenses (including consumption token)

The reconciliation loop does the following on each iteration:
- Extend the starter pack license
- Obtain node metrics from the collector
- For each product
  - determine how many 10 packs should be checked out
  - check that number against the current checked out licenses for this product
  - if additional capacity is needed
    - list licenses for this product and check out one that isn't being used
    - if they are all being used, customer needs more licenses and compliance status should be updated
  - if they are over-licensed, e.g. have 5x10 node licenses and < 40 nodes
    - check in a license for this product
  - otherwise, no action is needed

#### Notes

If any of the above checkout operations fail, the manager should block until they succeed. 
This is because the checkout type is `PROVISIONAL`, so AWS will check the licenses in automatically after an hour.
A license can only be checked out by one entity at any given point in time.

If there are insufficient licenses for any of the checkout operations, the api should return an informative error message about the customer's current license compliance

#### License Manager API Calls

1. List Received Licenses
- https://docs.aws.amazon.com/license-manager/latest/APIReference/API_ListReceivedLicenses.html

request body:
```
{
   "Filters": [ 
      { 
         "Name": "string", // Beneficiary, ProductSKU
         "Values": [ "string" ]
      }
   ],
   "LicenseArns": [ "string" ],
   "MaxResults": number,
   "NextToken": "string"
}
```
None of these fields are required

Need IAM permissions for licenses specific to Rancher products

2. Checkout License
- https://docs.aws.amazon.com/license-manager/latest/APIReference/API_CheckoutLicense.html

request body:
```
{
   "Beneficiary": "string",
   "CheckoutType": "string", // PROVISIONAL
   "ClientToken": "string", // generated token
   "Entitlements": [ 
      { 
         "Name": "string",
         "Unit": "string",
         "Value": "string"
      }
   ],
   "KeyFingerprint": "string",
   "NodeId": "string",
   "ProductSKU": "string"
}
```

`CheckoutType, Entitlements, KeyFingerprint, and ProductSKU` are all required fields

3. Extend License
- https://docs.aws.amazon.com/license-manager/latest/APIReference/API_ExtendLicenseConsumption.html

request body:
```
{
  "DryRun": boolean,
  "LicenseConsumptionToken": "string"
}
```

`LicenseConsumptionToken` is a required field

4. Checkin License
- https://docs.aws.amazon.com/license-manager/latest/APIReference/API_CheckInLicense.html

request body:
```
{
   "Beneficiary": "string",
   "LicenseConsumptionToken": "string"
}
```

`LicenseConsumptionToken` is a required field

5. Get License
- https://docs.aws.amazon.com/license-manager/latest/APIReference/API_GetLicense.html

request body:
```
{
   "LicenseArn": "string",
   "Version": "string"
}
```

`LicenseArn` is a required field


## Packages

- `pkg/metrics` - handles scraping, parsing and filtering of metrics from Rancher's `/metrics` endpoint
- `pkg/server` - simple server for the app
- `pkg/marketplace` - CSP marketplace wrappers
- `pkg/supportconfig`- Contains generation logic for SUSE CSP Support Config