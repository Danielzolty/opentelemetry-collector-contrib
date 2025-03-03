# Resource Detection Processor

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [beta]: traces, metrics, logs   |
| Distributions | [contrib], [aws], [observiq], [redhat], [splunk], [sumo] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aprocessor%2Fresourcedetection%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aprocessor%2Fresourcedetection) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aprocessor%2Fresourcedetection%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aprocessor%2Fresourcedetection) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@Aneurysm9](https://www.github.com/Aneurysm9), [@dashpole](https://www.github.com/dashpole) |

[beta]: https://github.com/open-telemetry/opentelemetry-collector#beta
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[aws]: https://github.com/aws-observability/aws-otel-collector
[observiq]: https://github.com/observIQ/observiq-otel-collector
[redhat]: https://github.com/os-observability/redhat-opentelemetry-collector
[splunk]: https://github.com/signalfx/splunk-otel-collector
[sumo]: https://github.com/SumoLogic/sumologic-otel-collector
<!-- end autogenerated section -->

The resource detection processor can be used to detect resource information from the host,
in a format that conforms to the [OpenTelemetry resource semantic conventions](https://github.com/open-telemetry/semantic-conventions/tree/main/docs/resource), and append or
override the resource value in telemetry data with this information.

## Supported detectors

### Environment Variable

Reads resource information from the `OTEL_RESOURCE_ATTRIBUTES` environment
variable. This is expected to be in the format `<key1>=<value1>,<key2>=<value2>,...`, the
details of which are currently pending confirmation in the OpenTelemetry specification.

Example:

```yaml
processors:
  resourcedetection/env:
    detectors: [env]
    timeout: 2s
    override: false
```

### System metadata

Note: use the Docker detector (see below) if running the Collector as a Docker container.

Queries the host machine to retrieve the following resource attributes:

    * host.arch
    * host.name
    * host.id
    * host.ip
    * host.cpu.vendor.id
    * host.cpu.family
    * host.cpu.model.id
    * host.cpu.model.name
    * host.cpu.stepping
    * host.cpu.cache.l2.size
    * os.description
    * os.type

By default `host.name` is being set to FQDN if possible, and a hostname provided by OS used as fallback.
This logic can be changed with `hostname_sources` configuration which is set to `["dns", "os"]` by default.

Use the following config to avoid getting FQDN and apply hostname provided by OS only:

```yaml
processors:
  resourcedetection/system:
    detectors: ["system"]
    system:
      hostname_sources: ["os"]
```

* all valid options for `hostname_sources`:
    * "dns"
    * "os"
    * "cname"
    * "lookup"

#### Hostname Sources

##### dns

The "dns" hostname source uses multiple sources to get the fully qualified domain name. First, it looks up the
host name in the local machine's `hosts` file. If that fails, it looks up the CNAME. Lastly, if that fails,
it does a reverse DNS query. Note: this hostname source may produce unreliable results on Windows. To produce
a FQDN, Windows hosts might have better results using the "lookup" hostname source, which is mentioned below.

##### os

The "os" hostname source provides the hostname provided by the local machine's kernel.

##### cname

The "cname" hostname source provides the canonical name, as provided by net.LookupCNAME in the Go standard library.
Note: this hostname source may produce unreliable results on Windows.

##### lookup

The "lookup" hostname source does a reverse DNS lookup of the current host's IP address.

### Docker metadata

Queries the Docker daemon to retrieve the following resource attributes from the host machine:

    * host.name
    * os.type

You need to mount the Docker socket (`/var/run/docker.sock` on Linux) to contact the Docker daemon.
Docker detection does not work on macOS.

Example:

```yaml
processors:
  resourcedetection/docker:
    detectors: [env, docker]
    timeout: 2s
    override: false
```

### Heroku metadata

When [Heroku dyno metadata is active](https://devcenter.heroku.com/articles/dyno-metadata), Heroku applications publish information through environment variables.

We map these environment variables to resource attributes as follows:

| Dyno metadata environment variable | Resource attribute                  |
|------------------------------------|-------------------------------------|
| `HEROKU_APP_ID`                    | `heroku.app.id`                     |
| `HEROKU_APP_NAME`                  | `service.name`                      |
| `HEROKU_DYNO_ID`                   | `service.instance.id`               |
| `HEROKU_RELEASE_CREATED_AT`        | `heroku.release.creation_timestamp` |
| `HEROKU_RELEASE_VERSION`           | `service.version`                   |
| `HEROKU_SLUG_COMMIT`               | `heroku.release.commit`             |

For more information, see the [Heroku cloud provider documentation](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/cloud-provider/heroku.md) under the [OpenTelemetry specification semantic conventions](https://github.com/open-telemetry/semantic-conventions).

```yaml
processors:
  resourcedetection/heroku:
    detectors: [env, heroku]
    timeout: 2s
    override: false
```

### GCP Metadata

Uses the [Google Cloud Client Libraries for Go](https://github.com/googleapis/google-cloud-go)
to read resource information from the [metadata server](https://cloud.google.com/compute/docs/storing-retrieving-metadata) and environment variables to detect which GCP platform the
application is running on, and detect the appropriate attributes for that platform. Regardless
of the GCP platform the application is running on, use the gcp detector:

Example:

```yaml
processors:
  resourcedetection/gcp:
    detectors: [env, gcp]
    timeout: 2s
    override: false
```

#### GCE Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_compute_engine")
    * cloud.account.id (project id)
    * cloud.region  (e.g. us-central1)
    * cloud.availability_zone (e.g. us-central1-c)
    * host.id (instance id)
    * host.name (instance name)
    * host.type (machine type)
    * (optional) gcp.gce.instance.hostname
    * (optional) gcp.gce.instance.name

#### GKE Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_kubernetes_engine")
    * cloud.account.id (project id)
    * cloud.region (only for regional GKE clusters; e.g. "us-central1")
    * cloud.availability_zone (only for zonal GKE clusters; e.g. "us-central1-c")
    * k8s.cluster.name
    * host.id (instance id)
    * host.name (instance name; only when workload identity is disabled)

One known issue is when GKE workload identity is enabled, the GCE metadata endpoints won't be available, thus the GKE resource detector won't be
able to determine `host.name`. In that case, users are encouraged to set `host.name` from either:
- `node.name` through the downward API with the `env` detector
- obtaining the Kubernetes node name from the Kubernetes API (with `k8s.io/client-go`)

#### Google Cloud Run Services Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_cloud_run")
    * cloud.account.id (project id)
    * cloud.region (e.g. "us-central1")
    * faas.id (instance id)
    * faas.name (service name)
    * faas.version (service revision)

#### Cloud Run Jobs Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_cloud_run")
    * cloud.account.id (project id)
    * cloud.region (e.g. "us-central1")
    * faas.id (instance id)
    * faas.name (service name)
    * gcp.cloud_run.job.execution ("my-service-ajg89")
    * gcp.cloud_run.job.task_index ("0")

#### Google Cloud Functions Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_cloud_functions")
    * cloud.account.id (project id)
    * cloud.region (e.g. "us-central1")
    * faas.id (instance id)
    * faas.name (function name)
    * faas.version (function version)

#### Google App Engine Metadata

    * cloud.provider ("gcp")
    * cloud.platform ("gcp_app_engine")
    * cloud.account.id (project id)
    * cloud.region (e.g. "us-central1")
    * cloud.availability_zone (e.g. "us-central1-c")
    * faas.id (instance id)
    * faas.name (service name)
    * faas.version (service version)

### AWS EC2

Uses [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/aws/ec2metadata/) to read resource information from the [EC2 instance metadata API](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) to retrieve the following resource attributes:

    * cloud.provider ("aws")
    * cloud.platform ("aws_ec2")
    * cloud.account.id
    * cloud.region
    * cloud.availability_zone
    * host.id
    * host.image.id
    * host.name
    * host.type

It also can optionally gather tags for the EC2 instance that the collector is running on.
Note that in order to fetch EC2 tags, the IAM role assigned to the EC2 instance must have a policy that includes the `ec2:DescribeTags` permission.

EC2 custom configuration example:
```yaml
processors:
  resourcedetection/ec2:
    detectors: ["ec2"]
    ec2:
      # A list of regex's to match tag keys to add as resource attributes can be specified
      tags:
        - ^tag1$
        - ^tag2$
        - ^label.*$
```

If you are using a proxy server on your EC2 instance, it's important that you exempt requests for instance metadata as [described in the AWS cli user guide](https://github.com/awsdocs/aws-cli-user-guide/blob/a2393582590b64bd2a1d9978af15b350e1f9eb8e/doc_source/cli-configure-proxy.md#using-a-proxy-on-amazon-ec2-instances). Failing to do so can result in proxied or missing instance data.

If the instance is part of AWS ParallelCluster and the detector is failing to connect to the metadata server, check the iptable and make sure the chain `PARALLELCLUSTER_IMDS` contains a rule that allows OTEL user to access `169.254.169.254/32`

### Amazon ECS

Queries the [Task Metadata Endpoint](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html) (TMDE) to record information about the current ECS Task. Only TMDE V4 and V3 are supported.

    * cloud.provider ("aws")
    * cloud.platform ("aws_ecs")
    * cloud.account.id
    * cloud.region
    * cloud.availability_zone
    * aws.ecs.cluster.arn
    * aws.ecs.task.arn
    * aws.ecs.task.family
    * aws.ecs.task.revision
    * aws.ecs.launchtype (V4 only)
    * aws.log.group.names (V4 only)
    * aws.log.group.arns (V4 only)
    * aws.log.stream.names (V4 only)
    * aws.log.stream.arns (V4 only)

Example:

```yaml
processors:
  resourcedetection/ecs:
    detectors: [env, ecs]
    timeout: 2s
    override: false
```

### Amazon Elastic Beanstalk

Reads the AWS X-Ray configuration file available on all Beanstalk instances with [X-Ray Enabled](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/environment-configuration-debugging.html).

    * cloud.provider ("aws")
    * cloud.platform ("aws_elastic_beanstalk")
    * deployment.environment
    * service.instance.id
    * service.version

Example:

```yaml
processors:
  resourcedetection/elastic_beanstalk:
    detectors: [env, elastic_beanstalk]
    timeout: 2s
    override: false
```

### Amazon EKS

    * cloud.provider ("aws")
    * cloud.platform ("aws_eks")
    * k8s.cluster.name

Note: The kubernetes cluster name is only available when running on EC2 instances, and requires permission to run the `EC2:DescribeInstances` [action](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html).
If you see an error with the message `context deadline exceeded`, please increase the timeout setting in your config.

Example:

```yaml
processors:
  resourcedetection/eks:
    detectors: [env, eks]
    timeout: 15s
    override: false
```

### AWS Lambda

Uses the AWS Lambda [runtime environment variables](https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html#configuration-envvars-runtime)
to retrieve the following resource attributes:

[Cloud semantic conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/cloud.md)

* `cloud.provider` (`"aws"`)
* `cloud.platform` (`"aws_lambda"`)
* `cloud.region` (`$AWS_REGION`)

[Function as a Service semantic conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/faas.md)
and [AWS Lambda semantic conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/aws-lambda.md)

* `faas.name` (`$AWS_LAMBDA_FUNCTION_NAME`)
* `faas.version` (`$AWS_LAMBDA_FUNCTION_VERSION`)
* `faas.instance` (`$AWS_LAMBDA_LOG_STREAM_NAME`)
* `faas.max_memory` (`$AWS_LAMBDA_FUNCTION_MEMORY_SIZE`)

[AWS Logs semantic conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/cloud-provider/aws/logs.md)

* `aws.log.group.names` (`$AWS_LAMBDA_LOG_GROUP_NAME`)
* `aws.log.stream.names` (`$AWS_LAMBDA_LOG_STREAM_NAME`)

Example:

```yaml
processors:
  resourcedetection/lambda:
    detectors: [env, lambda]
    timeout: 0.2s
    override: false
```

### Azure

Queries the [Azure Instance Metadata Service](https://aka.ms/azureimds) to retrieve the following resource attributes:

    * cloud.provider ("azure")
    * cloud.platform ("azure_vm")
    * cloud.region
    * cloud.account.id (subscription ID)
    * host.id (virtual machine ID)
    * host.name
    * azure.vm.name (same as host.name)
    * azure.vm.size (virtual machine size)
    * azure.vm.scaleset.name (name of the scale set if any)
    * azure.resourcegroup.name (resource group name)

Example:

```yaml
processors:
  resourcedetection/azure:
    detectors: [env, azure]
    timeout: 2s
    override: false
```

### Azure AKS

  * cloud.provider ("azure")
  * cloud.platform ("azure_aks")

```yaml
processors:
  resourcedetection/aks:
    detectors: [env, aks]
    timeout: 2s
    override: false
```

### Consul

Queries a [consul agent](https://www.consul.io/docs/agent) and reads its' [configuration endpoint](https://www.consul.io/api-docs/agent#read-configuration) to retrieve the following resource attributes:

  * cloud.region (consul datacenter)
  * host.id (consul node id)
  * host.name (consul node name)
  * *exploded consul metadata* - reads all key:value pairs in [consul metadata](https://www.consul.io/docs/agent/options#_node_meta) into label:labelvalue pairs.

```yaml
processors:
  resourcedetection/consul:
    detectors: [env, consul]
    timeout: 2s
    override: false
```

### Heroku

** You must first enable the [Heroku metadata feature](https://devcenter.heroku.com/articles/dyno-metadata) on the application **

Queries [Heroku metadata](https://devcenter.heroku.com/articles/dyno-metadata) to retrieve the following resource attributes:

* heroku.release.version (identifier for the current release)
* heroku.release.creation_timestamp (time and date the release was created)
* heroku.release.commit (commit hash for the current release)
* heroku.app.name (application name)
* heroku.app.id (unique identifier for the application)
* heroku.dyno.id (dyno identifier. Used as host name)

```yaml
processors:
  resourcedetection/heroku:
    detectors: [env, heroku]
    timeout: 2s
    override: false
```

### K8S Node Metadata

Queries the K8S api server to retrieve the following node resource attributes:

    * k8s.node.uid

The following permissions are required:
```yaml
kind: ClusterRole
metadata:
  name: otel-collector
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list"]
```

| Name | Type | Required | Default         | Docs                                                                                                                                                                                                                                   |
| ---- | ---- |----------|-----------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| auth_type | string | No       | `serviceAccount` | How to authenticate to the K8s API server.  This can be one of `none` (for no auth), `serviceAccount` (to use the standard service account token provided to the agent pod), or `kubeConfig` to use credentials from `~/.kube/config`. |
| node_from_env_var | string | Yes      | `K8S_NODE_NAME` | The environment variable name that holds the name of the node to retrieve metadata from. Default value is `K8S_NODE_NAME`. You can set the env dynamically on the workload definition using the downward API; see example              |

#### Example using the default `node_from_env_var` option:

```yaml
processors:
  resourcedetection/k8snode:
    detectors: [k8snode]
```
and add this to your workload:
```yaml
        env:
          - name: K8S_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
```

#### Example using a custom variable `node_from_env_var` option:
```yaml
processors:
  resourcedetection/k8snode:
    detectors: [k8snode]
    k8snode:
      node_from_env_var: "my_custom_var"
```
and add this to your workload:
```yaml
        env:
          - name: my_custom_var
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
```

### Openshift

Queries the OpenShift and Kubernetes API to retrieve the following resource attributes:

    * cloud.provider
    * cloud.platform
    * cloud.region
    * k8s.cluster.name

The following permissions are required:
```yaml
kind: ClusterRole
metadata:
  name: otel-collector
rules:
- apiGroups: ["config.openshift.io"]
  resources: ["infrastructures", "infrastructures/status"]
  verbs: ["get", "watch", "list"]
```

By default, the API address is determined from the environment variables `KUBERNETES_SERVICE_HOST`, `KUBERNETES_SERVICE_PORT` and the service token is read from `/var/run/secrets/kubernetes.io/serviceaccount/token`.
If TLS is not explicit disabled and no `ca_file` is configured `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` is used.
The determination of the API address, ca_file and the service token is skipped if they are set in the configuration.

Example:

```yaml
processors:
  resourcedetection/openshift:
    detectors: [openshift]
    timeout: 2s
    override: false
    openshift: # optional
      address: "https://api.example.com"
      token: "token"
      tls:
        insecure: false
        ca_file: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
```

See: [TLS Configuration Settings](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md) for the full set of available options.

## Configuration

```yaml
# a list of resource detectors to run, valid options are: "env", "system", "gce", "gke", "ec2", "ecs", "elastic_beanstalk", "eks", "lambda", "azure", "heroku", "openshift"
detectors: [ <string> ]
# determines if existing resource attributes should be overridden or preserved, defaults to true
override: <bool>
# [DEPRECATED] When included, only attributes in the list will be appended.  Applies to all detectors.
attributes: [ <string> ]
```

Moreover, you have the ability to specify which detector should collect each attribute with `resource_attributes` option. An example of such a configuration is:

```yaml
resourcedetection:
  detectors: [system, ec2]
  system:
    resource_attributes:
      host.name:
        enabled: true
      host.id:
        enabled: false
  ec2:
    resource_attributes:
      host.name:
        enabled: false
      host.id:
        enabled: true
```

### Migration from attributes to resource_attributes

The `attributes` option is deprecated and will be removed soon, from now on you should enable/disable attributes through `resource_attributes`.
For example, this config:

```yaml
resourcedetection:
  detectors: [system]
  attributes: ['host.name', 'host.id']
```

can be replaced with:

```yaml
resourcedetection:
  detectors: [system]
  system:
    resource_attributes:
      host.name:
        enabled: true
      host.id:
        enabled: true
      os.type:
        enabled: false
```

## Ordering

Note that if multiple detectors are inserting the same attribute name, the first detector to insert wins. For example if you had `detectors: [eks, ec2]` then `cloud.platform` will be `aws_eks` instead of `ec2`. The below ordering is recommended.

### GCP

* gke
* gce

### AWS

* lambda
* elastic_beanstalk
* eks
* ecs
* ec2

The full list of settings exposed for this extension are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).
