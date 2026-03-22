//go:build !(pacific || quincy) && ceph_preview

/*
Package builder (aka "commands builder") contains a set of
functions and types for working with Ceph's dynamic command framework
dynamically.

# Ceph Command Basics

Ever wonder how the Ceph project's `ceph` command works? Commands like `ceph
osd df`, `ceph pg ls`, or even `ceph orch ls`?  Ceph's core protocol is RADOS
and Ceph's RADOS library provides API functions such as `rados_mon_command`,
`rados_mgr_command`, `rados_osd_command` and so on. The go-ceph rados package
implements wrapper functions like `MonCommand` and `MgrCommand`. These API
functions serve as the fundamental building-blocks of Ceph's command line
interface.

As an example the function signature of MonCommand from the rados library
is as follows:

	func (c *Conn) MonCommand(args []byte) ([]byte, string, error) {
	  ...
	}

The input to this function is a bytes slice. The output is a byte slice
representing the server's response, a status string, and an error. The error
can be a protocol error or an error response from the server.

To invoke a particular command API on the server the input bytes slice must
contain a JSON formatted object containing the command and any parameters. The
command itself is formatted as a single space-separated string using the
"prefix" key and other parameters are a passed as keys and values in the JSON
object. For example the command `ceph osd df tree --filter=hdd` is equivalent
to the following JSON:

	{
	  "prefix": "osd df",
	  "output_method": "tree",
	  "filter": "hdd"
	}

If you know what parameters are expected it's often easy to construct some
static JSON or use a `map[string]string` and `json.Marshal` to create a
parameterized call.

Knowing what commands exist and what input variables are available is where
this package comes in. Ceph provides an API to query what is available
`{"prefix": "get_command_descriptions"}` and this package aims to make that
more convenient to use.

Note that the output of the commands is highly dependent on the command being
called. Some commands emit human readable text, others JSON, etc.  In some
cases, the format of the command's output can be requested by specifying a
"format" key and value such as "json" or "yaml". However, a command may not
support a particular format and ignore the hint.  Also, unlike the command
descriptions Ceph doesn't provide general structured descriptions of the
returned values even when emitting a machine-parseable format like JSON. You
may want to keep the documentation handy for that phase.

# Introducing go-ceph's Command Builder Package

The go-ceph library already provides many packages that wrap Ceph commands such
as `rbd/admin`, `cephfs/admin`, `common/admin/nfs` and so on. Our convention is
to call these admin packages because the APIs are typically needed for
administering a Ceph cluster rather than just storing/retrieving data from it.

But there are cases where we have not covered a set of APIs with a dedicated
package or there might be cases were you want to do things differently.
This commands builder package allows you to use the rados APIs to query
for the command descriptions and optionally use those descriptions to build
the command JSON.

# Querying Command Descriptions

This library currently provides two sets of APIs for querying command descriptions.
For querying the Ceph MON:

	QueryMonJSON(m ccom.MonCommander) ([]byte, error)
	QueryMonDescriptions(m ccom.MonCommander) (CommandDescriptions, error)

For querying the Ceph MGR:

	QueryMgrJSON(m ccom.MgrCommander) ([]byte, error)
	QueryMgrDescriptions(m ccom.MgrCommander) (CommandDescriptions, error)

The functions ending in ...JSON always return the raw JSON text in case
you want to do custom parsing or perhaps just want to dump the unedited
JSON to the output. The Descriptions functions will parse the JSON
automatically and return a helper type that stores the descriptions and
provides methods for searching for matching commands.

For example to query the MON for commands starting with "osd" one can execute:

	cde, err := QueryMonDescriptions(radosConn)
	// handle err...
	for _, cmd := range cde.Find("osd") {
	  fmt.Printf("osd command: %s\n", cmd.PrefixString())
	}

Each command description contains a signature that can be broken down
into the fixed prefix strings and the variables. Each [SignatureVariable]
contains fields that describe what type of input is expected, and sometimes
a bit about what the allowed values are.

# Building Commands

In addition to simply getting information about the commands and their
arguments the [Builder] type and Ceph argument types (those matching the
CephArgumentType interface) can be used to dynamically construct a
Values map that will be serialized to JSON. The `Apply` function can
be used to convert a sequence of strings and/or a mapping of keyword-value
pairs to a ceph command.

NOTE: Not all types are fully implemented. Consider filing an issue or
contributing a patch if you need one.

# Non-Goals

Note that this package just aims to provide components one can use to
dynamically construct Ceph command inputs without doing it all yourself. It
doesn't aim to be a replacement for the `ceph` command. It doesn't aim to be a
fully featured tool for building an alternative command line parser for ceph.
One might use it as a component of such a thing, or maybe a GUI, but that's all
it aims to be - a toolkit rather than a complete solution.
*/
package builder
