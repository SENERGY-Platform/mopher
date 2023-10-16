# what does it do?
mopher scans a github org to:
- inform about missing go version updates 
  - the go version of the mopher build will be used as the "latest" go version
  - patch version will be ignored. only major and minor version differences will be checked.
- warns if an org repository uses an old version of another repository of this org
  - only the master/main branch is checked
  - if the go.mod file uses a semantic version, the comparison uses the newest semantic version of the dependency
  - if the go.mod file uses a commit hash as version, the comparison uses the newest commit hash in the master/main branch of the dependency
- warns if a dev branch is not in sync with the master/main branch
- warns if a module name doesn't match its GitHub url
- lists a recommended update order
- generate a dependency graph in plantuml (optional)
- lists where a given dependency is used in which version in this org (optional)

# Version-Checks

## explicit org and dep parameters
```
mopher -org=SENERGY-Platform -dep=github.com/SENERGY-Platform/foo
mopher -org=SENERGY-Platform
mopher -org=SENERGY-Platform -graph=dependencies.plantuml
```
- the org flag states which github org to scan. ths value is mandatory but may be derived from other inputs
- the dep flag is optional and states for which dependency the org is scanned, to list ist usage (using repository and version)
- the graph flag is optional and outputs a dependency graph

## parameters by github address
```
mopher github.com/SENERGY-Platform/foo
mopher github.com/SENERGY-Platform
```
if the arg is a github url, it will be interpreted as the 'org' and if possible 'dep' flag


## parameters by project location
```
mopher /path/to/repo
```
mopher will get the 'org' and 'dep' flags from the go.mod file of the referenced dir


## parameters by local project (mopher in path variable) 
### assumptions
- mopehr is in "paths"
- the current dir contains a go.mod file
### call
```
mopher
```
### what happens
mopher will get the 'org' and 'dep' flags from the go.mod file of the current dir

## parameters by local project
### assumptions
- mopehr is not in "pahts"
- the current dir contains a go.mod file
### call
```
/path/to/mopher
```
### what happens
mopher will get the 'org' and 'dep' flags from the go.mod file of the current dir

# Output
the 'output' argument decides where the resulting warnings should be sent to:
- default: std-out, if nothing is set
- http (slack webhook), if with 'http://' or 'https://' prefix
- file location, is neither empty nor http

# Cron
the 'cron' lets mopher run repeatedly.
```
mopher -distinct -cron="* * * * *" github.com/SENERGY-Platform
```
the 'distinct' flag is optional and prevents repeated outputs of the same warnings 

# Graph
```
mopher -graph=graph.plantuml github.com/SENERGY-Platform
```
creates a file named "graph.plantuml" containing a plantuml description usable in http://www.plantuml.com/plantuml/uml to generate a UML-Component-Diagram

by default the command limits the nodes to org repositories. if all dependencies are wanted in the graph, use the graph_verbose flag
```
mopher -graph=graph.plantuml -graph_verbose github.com/SENERGY-Platform
```
WARNING: http://www.plantuml.com/plantuml/uml has a size limit

# Debugging
```
mopher -debug github.com/SENERGY-Platform
```

# Env
all program-arguments can be passed as environment variable. the envorionment variable name is the argument in caps lock camel-case with the 'MOPHER_' prefix. for example 'output' becomes 'MOPHER_OUTPUT'

# Error-Handling
to handle
```
time=2023-08-28T07:37:47.838+02:00 level=INFO msg="error creating SSH agent: \"Error connecting to SSH_AUTH_SOCK: dial unix /run/user/1001/keyring/ssh: connect: resource temporarily unavailable\""
```
you can use the max_conn flag to change the max parallel git request count from the default 25
```
mopher -max_conn=10 github.com/SENERGY-Platform
```