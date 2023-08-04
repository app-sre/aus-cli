# Advanced Upgrade Service Command Line Tools

This project contains the `ocm aus` OCM CLI plugin, that simplifies the interaction with the Advanced Upgrade Service SRE capability available at <https://source.redhat.com/groups/public/sre/wiki/advanced_upgrade_service_aus>

## Installation

## Option 1

Use one of the releases from <https://github.com/app-sre/aus-cli/releases> or directly download the latest release for your architecture with

```shell
curl -L -o ocm-aus "https://github.com/app-sre/aus-cli/releases/latest/download/ocm_aus_$(uname -s)_$(uname -m)"
```

and move it into one of your `$PATH`s.

## Option 2

Execute the following go install command to build and install the AUS CLI into $GOPATH/bin.

```shell
go install github.com/app-sre/aus-cli/cmd/ocm-aus@latest
```

## Option 3

Alternatively you can build from sources with `make build`.

## Usage

Install the [ocm-cli](https://github.com/openshift-online/ocm-cli#installation) and ensure that you are able to log in to an active OCM environment.

Run `ocm login ...` to establish a login session with your chosen OCM environment. Login instructions can be found on <https://console.redhat.com/openshift/token>

Once logged in, the plug-in can be accessed by running `ocm aus` which will display command information and a command overviews.

## Manage cluster upgrade policies

Create a new cluster upgrade policy with `ocm aus apply policies [flags] [args]`

| Flag               | Definition                                                                                                                                         |
|--------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| --cluster-name     | Name of the cluster to manage a policy for. This name needs to match the cluster name in OCM.                                                      |
| --org-id           | The OCM organization ID the cluster lives in. Defaults to the organization ID of the currently logged in user.                                     |
| --schedule         | A cron expression that defines when the cluster should be upgraded or a schedule preset (weekdays, anytime)                                        |
| --workload         | An identifier for the workload that runs on the cluster. Soak days are calculated per workload. Can be specified multiple times.                   |
| --soak-days        | The number of days to wait before upgrading the cluster. Soak days are accumulated per version and workload within on organization. Defaults to 0. |
| --mutex            | The mutexs the cluster must hold before it can start an upgrade. Can be specified multiple times.                                                  |
| --sector           | The sector the cluster belongs to. Can be used to gate cluster upgrades between sets of clusters.                                                  |
| --blocked-versions | Blocked version expressions                                                                                                                        |
| --dry-run          | Test the command without taking any action.                                                                                                        |
| --dump             | Instead of applying the policy to the cluster, it is written to stdout in JSON format                                                              |

Policies can also be written to a file and applied from a file.

```shell
ocm aus apply policies --cluster-name my-cluster --workload service --schedule weekdays --dump | tee policy.json
[
  {
    "conditions": {
      "soak_days": 0
    },
    "name": "my-cluster",
    "schedule": "* * * * 1-4",
    "workloads": [
       "service"
    ]
  }
]

cat policy.json | ocm aus apply policies -
Apply cluster upgrade policy to my-cluster
```

The policy file can also contain multiple policies.

## Manage blocked versions

Versions can be blocked on an OCM organization level. The `version-blocks` sub-command can be used to block and unblock versions patterns. Patterns are specified as regular expressions.

Manage blocked versions with `ocm aus apply version-blocks [fags]`

| Flags              | Definition                                                                                                                     |
|--------------------|--------------------------------------------------------------------------------------------------------------------------------|
| --block-version    | Blocks a version. Can be specified multiple times.                                                                             |
| --unblock-version  | Remove a version block. Can be specified multiple times.                                                                       |
| --replace          | Replace all currently blocked versions with the ones specified by `--block-version`                                            |
| --org-id           | The OCM organization ID where the version blocks are managed. Defaults to the organization ID of the currently logged in user. |

```shell
ocm aus apply version-blocks --block-version "^4\\.13\\..*$" --block-version "^4\\.14\\..*$"
ocm aus apply version-blocks --unblock-version "^4\\.13\\..*$"

ocm aus get version-blocks
[
   "^4\\.13\\..*$"
]
```

Version blocks can also be written to a file and applied from a file.

```shell
ocm aus apply version-blocks --block-version "^4\\.13\\.1$" --block-version "^4\\.14\\..*$" --replace --dump | tee version-blocks.json
[
   "^4\\.13\\.1$"
   "^4\\.14\\..*$"
]

cat version-blocks.json | ocm aus apply version-blocks - --replace
Apply blocked versions to organization 2Q0awarcxlarxaWwrFFpbLITiGu
```

Together with the `--replace-versions` option, applying from a file makes sure that the desired state defined in the file is going to be the exact state on the organization.

## Manage sector dependencies

Sectors are dependant groups of clusters. A version is only considered for upgrade within a sector if all dependant sectors have been fully upgraded to to that version.

Create or replace an organizations sector dependencies with `ocm aus apply sectors [flags]`

| Flags        | Definition                                                                                                                               |
|--------------|------------------------------------------------------------------------------------------------------------------------------------------|
| --add-dep    | `A=B,C` ... Establishes a dependency from sector A to sectors B and C. Can be specified multiple times.                                  |
| --remove-dep | `A=B`   ... Deletes the dependency from sector A to sector B.                                                                            |
| --replace    | Replaces all existing sector dependencies with the ones specified by `--add-sector-dep` or the ones provided via stdin.                  |
| --org-id     | The OCM organization ID where the sectors and dependencies are defined. Defaults to the organization ID of the currently logged in user. |

```shell
ocm aus apply sectors --add-dep prod=stage --add-dep stage=dev,dev-2
ocm aus apply sectors --remove-dep stage=dev-2
Apply sector configuration to organization 2Q0awarcxlarxaWwrFFpbLITiGu

ocm aus get sectors
[
  {
    "dependencies": [
       "dev"
    ],
    "name": "stage"
  },
  {
    "dependencies": [
       "stage"
    ],
    "name": "prod"
  }
]
```

Sector dependencies can also be written to a file and applied from a file.

```shell
ocm aus apply sectors --add-dep prod=stage --add-dep stage=dev --replace --dump | tee sector-deps.json
[
  {
    "dependencies": [
       "dev"
    ],
    "name": "stage"
  },
  {
    "dependencies": [
       "stage"
    ],
    "name": "prod"
  }
]

cat sector-deps.json | ocm aus apply sectors - --replace
Apply sector configuration to organization 2Q0awarcxlarxaWwrFFpbLITiGu
```

Together with the `--replace-sector-deps` option, applying from a file makes sure that the desired state defined in the file is going to be the exact state on the organization.

## Example

We will create policies for two stage and two production clusters. We want them to upgrade as follows:
- stage-1 cluster is upgraded immediately to any new version
- stage-2 cluster is upgraded a day later
- only one stage cluster is upgraded at a time
- once all stage clusters are upgraded and a version has soak for 5 days, prod-1 and prod-2 are upgrade
- only one production cluster is upgraded at a time
- upgrades to version 4.13 should be blocked

First create the policies for the stage clusters. `stage-1` defines `0` soak days, so an upgrade is scheduled immediately for every new version. `stage-2` cluster defines `1` soak day, therefore a another cluster with the same workload (`stage-1`) must run with a version for `1` day before the upgrade is scheduled. Both clusters share the same mutex, so only one of them can upgrade at a time.

```shell
ocm aus apply policies \
  --cluster-name stage-1 \
  --schedule weekdays \
  --workload my-service \
  --sector stage \
  --mutex stage-mutex \
  --soak-days 0

ocm aus apply policies \
  --cluster-name stage-2 \
  --schedule weekdays \
  --workload my-service \
  --sector stage \
  --mutex stage-mutex \
  --soak-days 1
```

Both clusters belong to the `stage` sector. We will see in a moment how we can make sure that all stage clusters must be upgraded to a version before it is considered for the production clusters. First, lets create the policies for the production clusters.

```shell
ocm aus apply policies \
  --cluster-name prod-1 \
  --schedule weekdays \
  --workload my-service \
  --sector prod \
  --mutex prod-mutex \
  --soak-days 5

ocm aus apply policies \
  --cluster-name prod-2 \
  --schedule weekdays \
  --workload my-service \
  --sector prod \
  --mutex prod-mutex \
  --soak-days 5
```

They share their own mutex `prod-mutex` so only one production cluster is upgraded at a time. Both of them define `5` soak days, so a version must have soaked for 5 days on other clusters with the same workload, so on the stage clusters.

We want to make sure that the production clusters upgrade only after ALL stage clusters have been upgraded. Soak days are not sufficient to ensure this condition, because a version could potentially also soak long enough on only one of the stage clusters.

Let's define that the `prod` sector depends on the `stage` sector.

```shell
ocm aus apply sectors --add-dep prod=stage

ocm aus get sectors
[
  {
    "dependencies": [
       "stage"
    ],
    "name": "prod"
  }
]
```

Now let's make sure upgrades to 4.13.x are blocked.

```shell
ocm aus apply version-blocks --block-version "^4\\.13\\..*$"

ocm aus get version-blocks
[
   "^4\\.13\\..*$"
]
```

With `ocm aus status` you can inspect the entire configuration

```shell
Organization ID:           2Q0awarcxlarxaWwrFFpbLITiGu
Organization name:         My Organization
OCM environment:           https://api.openshift.com
Blocked Versions:          ^4\.13\..*$
Sector Configuration:      (2 in total)
  Name                     Depends on
  ----                     ----------
  prod                     stage
  stage
Clusters:  (4 in total)
  Cluster Name             AUS enabled  Schedule     Sector  Mutexes      Soak Days  Workloads
  ------------             -----------  --------     ------  -------      ---------  ---------
  stage-1                  true         * * * * 1-4  stage   stage-mutex  0          my-service
  stage-2                  true         * * * * 1-4  stage   stage-mutex  1          my-service
  prod-1                   true         * * * * 1-4  prod    prod-mutex   5          my-service
  prod-2                   true         * * * * 1-4  prod    prod-mutex   5          my-service
```

