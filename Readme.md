# what does it do?
mopher scans a github org to:
- inform about missing go version updates (the go version of the mopher build will be used)
- warns if an org repository is used as a dependency in another org repository (main/master branch) with a different version than the newest main/master commit
- generate a dependency graph in plantuml (optional)
- lists where a given dependency is used in which version in this org (optional)

# 1
```
mopher -org=SENERGY-Platform -dep=github.com/SENERGY-Platform/foo
mopher -org=SENERGY-Platform
mopher -org=SENERGY-Platform -graph=dependencies.plantuml
```
- the org flag states which github org to scan. ths value is mandatory but may be derived from other inputs
- the dep flag is optional and states for which dependency the org is scanned, to list ist usage (using repository and version)
- the graph flag is optional and outputs a dependency graph

# 2
```
mopher github.com/SENERGY-Platform/foo
mopher github.com/SENERGY-Platform
```
if the arg is a github url, it will be interpreted as the 'org' and if possible 'dep' flag


# 3
```
mopher /path/to/repo
```
mopher will get the 'org' and 'dep' flags from the go.mod file of the referenced dir


# 4 
### assumptions
- mopehr is in "paths"
- the current dir contains a go.mod file
### call
```
mopher
```
### what happens
mopher will get the 'org' and 'dep' flags from the go.mod file of the current dir

# 5
### assumptions
- mopehr is not in "pahts"
- the current dir contains a go.mod file
### call
```
/path/to/mopher
```
### what happens
mopher will get the 'org' and 'dep' flags from the go.mod file of the current dir
