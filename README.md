# appman

A TUI tools that simple process manager for developers.

![appman screenshot](./misc/appman.png)

# Install

```shell
$ go install github.com/mrtc0/appman
```

# Configuration

### Example

```yaml
- name: "frontend"
  path: "/path/to/frontend"
  startCommand: ["npm", "run", "start"]
  port: 3000
  url: "http://localhost:3000"

- name: "api"
  path: "/path/to/backend"
  startCommand: ["go", "run", "main.go"]
  port: 5600
  url: "http://localhost:5600"

- name: "database"
  path: "/path/to/backend"
  startCommand: ["docker", "compose", "up"]
  stopCommand: ["docker", "compose", "down"]
```

### `name` (required)

`name` is application name.

### `path` (required)

`path` is the directory from which to run the application.

### `startCommand` (required)

`startCommand` is a command to start an application.

### `stopCommand` (optional)

`startCommand` is a command to stop an application.  
If `stopCommand` is not specified, appman will send SIGINT to the process when it stops the application.

### `port` (optional)

`port` is application port number.

### `url` (optional)

`url` is application URL.

# Command

- `Tab`: Switching focus.
- `Enter`: Item selection.
- `Ctrl + C` or `ESC`: Exit appman. All applications are automatically stopped.
