
### Historical Supported Ceph Versions

For recent releases, please refer to the [readme](../README.md).

The following tables describes what versions of Ceph were supported for
particular go-ceph releases. Note that prior to 2020 the project did not make
versioned releases. The ability to compile with a particular Ceph version
before go-ceph v0.2.0 is not guaranteed.

| go-ceph version | Supported Ceph Versions | Deprecated Ceph Versions |
| --------------- | ------------------------| -------------------------|
| v0.16.0         | octopus, pacific†       | nautilus                 |
| v0.15.0         | octopus, pacific        | nautilus                 |
| v0.14.0         | octopus, pacific        | nautilus                 |
| v0.13.0         | octopus, pacific        | nautilus                 |
| v0.12.0         | octopus, pacific        | nautilus                 |
| v0.11.0         | nautilus, octopus, pacific  |                      |
| v0.10.0         | nautilus, octopus, pacific  |                      |
| v0.9.0          | nautilus, octopus       |                          |
| v0.8.0          | nautilus, octopus       |                          |
| v0.7.0          | nautilus, octopus       |                          |
| v0.6.0          | nautilus, octopus       | mimic                    |
| v0.5.0          | nautilus, octopus       | luminous, mimic          |
| v0.4.0          | luminous, mimic, nautilus, octopus | |
| v0.3.0          | luminous, mimic, nautilus, octopus | |
| v0.2.0          | luminous, mimic, nautilus          | |
| (pre release)   | luminous, mimic  (see note)        | |

The tags affect what is supported at compile time. What version of the Ceph
cluster the client libraries support, and vice versa, is determined entirely
by what version of the Ceph C libraries go-ceph is compiled with.

† Preliminary support for Ceph Quincy was available, but not fully tested, in
this release.

