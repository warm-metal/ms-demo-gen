# MSDGen
MSDGen generates microservice demos of any given size and connectivity constraints.

```
msdgen | kubectl apply -f-
```

## MSD10(A generated demo with 10 services)

<img src="https://github.com/warm-metal/ms-demo-gen/blob/main/.res/topology.png?raw=true" />
<img src="https://github.com/warm-metal/ms-demo-gen/blob/main/.res/kiali.png?raw=true" width="700" />


- [Getting binary](#getting-binary)
- [Getting started](#getting-started)
- [Adjusting service behaviors](#adjusting-service-behaviors)
- [Changing size and connectivity](#changing-size-and-connectivity)
- [Operating MSDn](#operating-msdn)
- [Output options](#output-options)

## Getting binary

We have pre-built binaries for MacOS, Linux and Windows on amd64 architecture.
They can be found on the release page. 

If you need binaries for other OSes or architectures, feel free to file an issue.

## Getting started

Run `msdgen`, you will get a [MSD10](#msd10a-generated-demo-with-10-services)
which are manifests of 10 connected services and a traffic generator as their client can be deployed on K8s clusters.
The connectivity layout is saved in a DOT file named `connectivity-layout-foo.dot`.
It can be converted to an image by `dot -Tpng foo.dot -o bar.png` if you have graphviz installed.
If you have Istio and Kiali installed in cluster, 
you will see a graph has the same connectivity layout to which shown in the dot file.

Each MSDn has only 1 entry service, it is called `gateway`. Services are connected via HTTP protocol.
An MSDn also contains a client workload directly connects to the gateway service to generate traffic. Its name is `traffic-generator`.
You can increase the client bandwidth by starting more concurrent client processors,
or waiting less time between queries in each processor, through `-traffic-gen-proc` and `-traffic-gen-query-interval` respectively.

Every time you run `msdgen` will generate a new MSDn with various connectivity even with same parameters.
But, all Deployments and Services deployed in a cluster can be filtered with label `origin=msdgen`.
You can easily delete all of them.
```
kubectl delete svc,deploy --wait -l origin=msdgen
```

## Adjusting service behaviors
We provide a few options to coordinate service behaviors. This might enhance the similairty to real business.

* `-parallel` enables concurrent queries to all upstream services instead of fetching each of them sequentially.
* `-long` indicates that all queries to the same upstream are sent through the same L4 connection.
* `-timeout` is used to change the timeout of each upstream query.
* `-payload-size` sets the response body size of each downstream query.
* `-upload-size` changes the upstream query body size and its method. The query method will be GET if the options is set to 0. Otherwize, POST instead.

These options are shared by all services. Once you need to change only particular services,
corresponding environment variables can be found in generated manifests.

## Changing size and connectivity
MSDGen generates service connectivity randomly.
Specific connectivity definition for each service is not currently supported.
Even though, you can still rule the randomness of MSDGen.

* `-services` changes the total number of services.
* `-max-versions` defines the upperbound of the number of versions for each service.
* `-max-upstream` sets the upperbound of the number of upstreams for each service.
* `-max-downstream` sets the upperbound of the number of downstreams for each service.
* `-longest-call-chain` limits the number of services through which a query may walk.
If the topology of an MSDn can be a tree with the `gateway` service as its root, `-longest-call-chain` sets its height.

## Output options
MSDGen prints generated manifests to stdout by default.
If you'd like to change the output position, the `-o` option can be used to set the target directory.
Also, `-distributed-versions` defines the number of output files, and Versions of the same service will be distributed in those files randomly.
This will be helpful in a multi-cluster environment.

## Operating MSDn
MSDn is designed to use less CPU and memory, such that you can run a large scale of services on commodity computers.
As we tested, each servise costs less than 10MiB memory if both payload size and upload data size are less than 1MiB.
The CPU usage increases as the concurrent downstream size or the upstream size increasing, and finally less than 500m(1CPU=1000m).

We also provide two options to limit the CPU usage of each service, `-service-cpu-request` and `-service-cpu-limit`.
If you need to change the number of service workload replicas, `-max-replicas` will be its upperbound.
`-namespaces` is provided to randomly distribute services to different namespaces.
