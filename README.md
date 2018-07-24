# idetcd

## Name

*idetcd* - a etcd-based plugin used for identifying nodes in a cluster without domain name collsion.

## Description

## Syntax

~~~
idetcd {
	endpoint ENDPOINT...
	limit LIMIT
	pattern PATTERN
}
~~~

* `endpoint` **ENDPOINT** the etcd endpoints. Defaults to "http://localhost:2379".
* `limit` **LIMIT** the maximum limit of the node number in the cluster, if some nodes is going to expose itself after the node number in the cluster hits this limit, it will fail.
* `pattern` **PATTERN** the domain name pattern that every node follows in the cluster. And here we use golang template for the pattern.

## Examples

~~~ corefile
. {
	idetcd {
		endpoint http://localhost:2379
		limit 10
		pattern worker{{.ID}}.local.tf
	}
}

Multiple endpoints are supported as well, and pattern should follow a golang template format.
