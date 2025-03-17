# opereta

opereta is a lightweight remote task orchestration tool written in go. it uses ssh to run shell commands on remote hosts and supports task retries, inventory-based configuration, and custom task options.

## features

- run tasks concurrently on multiple hosts
- configure hosts with inventory files (including ssh retry options)
- set per-task options like retries, delay, and no_result (which displays a fixed message instead of command output)
- detailed logging with unique ids per task and host session
- output results as json or in a human-readable format

## installation

build the tool with go:

```sh
go build -o opereta ./cmd/main.go
```

## usage

opereta accepts the following flags:

- **-inventory**: path to the inventory file (default: `configs/inventory.yml`)
- **-tasks**: path to the tasks file (default: `configs/tasks.yml`)
- **-output-json**: output results in json format
- **-parallel**: run tasks concurrently (default: true)

example command:

```sh
./opereta -inventory=./configs/inventory.yml -tasks=./configs/tasks.yml
```

result:
<pre>
[4c0ed69a][c3088585] <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> test <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> Success echo hello
Executed At: 2025-03-17T14:45:07+01:00 | Duration: 573.487803ms
Result:
hello world

[4c0ed69a][30874cc8] <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> test <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> Success display date
Executed At: 2025-03-17T14:45:07+01:00 | Duration: 124.661992ms
Result:
Mon Mar 17 13:45:07 UTC 2025

[4c0ed69a][5cbfc9e8] <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> test <span style="color: rgba(0, 255, 34, 0.69);">[-]</span> Success echo test
Executed At: 2025-03-17T14:45:08+01:00 | Duration: 123.482879ms
Result:
test

[f9b0b944][957fa1a2] <span style="color: rgb(201, 60, 60);">[-]</span> err <span style="color: rgb(201, 60, 60);">[-]</span> Error Task echo hello failed: after 1 task attempts, SSH connection failed: dialing SSH after 3 attempts: dial tcp 10.0.4.2:2222: connect: connection refused
Executed At: 2025-03-17T14:45:14+01:00 | Duration: 6.045720843s

<span style="color: rgb(0, 153, 255);">Summary per server:</span>
<span style="color: rgba(0, 255, 34, 0.69);">Server: test - Total: 3, Success: 3, Failed: 0, Status: OK</span>
<span style="color: rgb(201, 60, 60);">Server: err - Total: 1, Success: 0, Failed: 1, Status: UNREACHABLE</span>
<span style="color: rgb(0, 153, 255);">All tasks completed.</span>
</pre>


## configuration

### inventory file

define your remote hosts in a yaml file. for example:

```yaml
hosts:
  - name: "monitoring"
    address: "10.0.4.50"
    port: 22
    user: "monitoring"
    private_key: "/home/username/.ssh/id_ed25519"
    retry_ssh: "3s"
    retry_ssh_count: 3
```

### tasks file

define your tasks in a yaml file. for example:

```yaml
- name: "check uptime"
  module: "shell"
  max_retries: 3
  retry_delay: "2s"
  params:
    command: "uptime"

- name: "show current user"
  module: "shell"
  no_result: true
  params:
    command: "whoami"
```

the `no_result` option, when set to true, makes opereta display a fixed message ("no result setup in task") instead of the actual command output.

## logging

opereta logs each task execution with a unique event id and a host-specific global id. these ids help you trace each task and host session. logs are printed in interactive mode and can also be output in json using the **-output-json** flag.

## contributions

feel free to contribute. open issues, pull requests, and suggestions are welcome.
please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## license

MIT licensed.
see [LICENSE](LICENSE) for full details.

## author

- [Lucian BLETAN](https://github.com/gni)