## upgrade-all-services (cf cli plugin)

A CF-CLI plugin for upgrading all service instances in a CF foundation.

### Purpose
This tool was developed to allow users to upgrade all service instances they have access to in a foundation, without having to navigate between orgs and spaces to discover upgradable instances.

**Warning:** It is important to ensure that the authenticated user only has access to instances which you wish to upgrade. The plugin will upgrade all service instances a user has access to, irrespective of org and space. 

### Installing
First build the binary from the plugin directory
```
go build .
```
Then install the plugin using the cf cli
```
cf install-plugin <path_to_plugin_binary>
```

### Releasing
To create a new GitHub release, decide on a new version number [according to Semantic Versioning](https://semver.org/), and then:
1. Create a tag on the main branch with a leading `v`:
   `git tag vX.Y.X`
1. Push the tag:
   `git push --tags`
1. Wait for the GitHub action to run GoReleaser and create the new GitHub release


### Usage

```
cf upgrade-all-services <broker_name> [options]

Options:
    -parallel                                 - number of upgrades to run in parallel (defaults to 10)
    -loghttp                                  - log HTTP requests and responses
    -dry-run                                  - print the service instances that would be upgraded
    -min-version-required <major.minor.patch> - checks and fails if any service instance has a version less than the minimum required <major.minor.patch>
    -fail-if-not-up-to-date                   - checks and fails if any service instance is not up-to-date. An instance is not up-to-date if it is marked as upgradable or belongs to a deactivated plan
    -check-deactivated-plans                  - checks and fails if any of the plans have been deactivated
```