# Installation

```shell
go install github.com/xh-dev-go/plantuml-watcher
```

# Usage
```shell
# Show what path will be added to watch
plantuml-watcher -showOnly -dir {directory}

# watch the path and subdirectory for any *.puml file and save the png and svg file
plantuml -dir {directory} 
```

# Design
![](./docs/flow.svg)