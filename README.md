# xenmatchbox

`xenmatchbox` is a tool to boot Xen PV domUs by using kernel, initrd and kernel arguments from [matchbox](https://github.com/coreos/matchbox).

Currently a work-in-progress but code is tested to work on Alpine Xen 4.8 dom0 with CoreOS 1465.7.0.

## Usage

### matchbox server

Run a [matchbox](https://github.com/coreos/matchbox) server with gRPC API enabled (the `-rpc-address` flag), which requires CA, server and client certificates present.

### Building `xenmatchbox`

Tested to build with Go 1.9, with dependencies using [dep](https://github.com/golang/dep).

For example:
```
dep ensure
GOOS=linux go build .
scp xenmatchbox user@dom0:/path/to/xenmatchbox
```

### domU configuration

Remove eventual `kernel`, `ramdisk` and `extra` parameters and provide the bootloader as:

```
bootloader = '/path/to/xenmatchbox'
bootloader_args = [
  '--server', 'matchbox.foo',
  '--httpport', '8080',
  '--grpcport', '8081',
  '--ca', '/path/to/ca.crt',
  '--cert', '/path/to/client.crt',
  '--key', '/path/to/client.key',
  '--lookup', 'mac=00:01:02:03:04:05',
]
```

This will perform a lookup of the matchbox profile based on the MAC address `00:01:02:03:04:05`. Probably all other keys used in the profile can be used for lookup.

*NOTE:* xl currently requires a disk to be configured for the domU to use the `bootloader` parameter...

### Argument templating

The current examples in the [matchbox](https://github.com/coreos/matchbox) works by using variables in the return args that are expanded by either iPXE or Grub.

Similar functionality exists in `xenmatchbox`. Using the [`text/template`](https://golang.org/pkg/text/template/) package, variables listed in `--lookup` are expanded.

Example:
```
"args": [
  "coreos.config.url=http://matchbox/ignition?mac={{ .mac }}"
]
```

### Start domU

With configuration as above, it should be as easy as running `xl create /path/to/domu.cfg`.

Note that `xenmatchbox` gets an output directory from `xl`, where it will save the kernel and initrd. Since CoreOS initrd is quite large, make sure enough space is provided. Typically this is somewhere under `/run`.
