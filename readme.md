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