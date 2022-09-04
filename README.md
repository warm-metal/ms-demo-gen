# ms-demo-gen(MSDGen)
MSDGen can generate microservice demos of any given scale of service number.
It creates k8s manifests in a yaml file and optionaly a DOT file which contains its topology data.

MSDGen is built for service mesh testing.

## Usage
We have pre-built binaries for MacOS, Linux and Windows on amd64 architecture.
You can download them in the release page. 
If you need binaries on other platforms and can't build by yourselves, feel free to file an issue.

```
msdgen
```
Run `msdgen`, you will get manifests of 10 connected services and a traffic generator as client.
All Deployments and Services can be listed with label `origin=msdgen`.

Each demo project has only 1 entry service, it is called `gateway`.

The topoloty chart can be generated by `dot -Tpng foo.dot -o bar.png` if you have graphviz installed.

### Change service behaviors
A few options are given to coordinate service behaviors. This might enhance the similairty to real business.

The upstream call style can be changed by `-parallel` and `-long`.
The former option enables concurrent queries to all upstream services,
and the latter one indicates that all queries to the same upstream are excecuted in the same L4 connection.
You can also change the query timeout through `-timeout`.

`-payload-size` and `-upload-size` is used to change actions of a single query.

These options are shared by all services. Once you need to change only particular services,
you can found the corresponding environment variables in the generated manifests.

### traffic-generator
The name of the client is `traffic-generator`.
You can increase the client bandwidth by starting more concurrent client processors,
or waiting less time between queries in each processor, through `-traffic-gen-proc` and `-traffic-gen-query-interval` respectively.

### Change scale and topology
MSDGen currently generates service connectivity randomly under given constraints.

You can specify the total number of services through `-services`.
Then, set how many upstreams a service can connect at most via `-max-upstream`
and a upper bound of downstream number of a service through `-max-downstream`.
`-longest-call-chain` can be used to limit the number of services which a query may walk through.
If the topology of a demo can be a tree with `gateway` service as its root, `-longest-call-chain` will define its height.

If you need to change the number of service workload replicas, `-max-replicas` will be its upper bound. `-namespaces` is provided to randomly distribute services to multiple namespaces.

## Service Benchmark
The service benchmark will help you determine the **CPU** and **memory** capacity a demo takes.

### A single processor client with empty payload queries and a service with various number of upstreams
| Upstream | CPU(m) | Memory(Mi) |
| --- | --- | --- |
|0| `66`|`7`|
|1| `272`|`7`|
|2| `523`|`7`|
|5| `500`|`8`|
|10| `542`|`8`|
|20| `571`|`9`|

### A client withou various processors with empty paylaod queries and a service w/o upstream
| Concurrent Query Processors | CPU(m) | Memory(Mi) |
| --- | --- | --- |
|1| `164`|`7`|
|2| `253`|`7`|
|5| `400`|`7`|
|10| `476`|`7`|
|20| `512`|`8`|

### A single processor client with nonempty queries and a service w/o upstream
| Upload Size(byte) | CPU(m) | Memory(Mi) |
| --- | --- | --- |
|256| `236`|`7`|
|512| `236`|`7`|
|1Ki| `229`|`7`|
|50Ki| `542`|`6`|
|1Mi| `505`|`6`|
|10Mi| `499`|`26`|

### A single processor client with nonempty queries and a service with 5 upstream services
| Payload Size(byte) | CPU(m) | Memory(Mi) |
| --- | --- | --- |
|256| `571`|`7`|
|512| `548`|`7`|
|1Ki| `553`|`7`|
|50Ki| `339`|`7`|
|1Mi| `197`|`5`|
|10Mi| `185`|`25`|