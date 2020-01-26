# Ownershit

Manage repository ownership with your organization's repositories.

## Usage

This tool expects a `yaml` file called `repositories.yaml` in the path you run this, 
but you can override this with the `--config` flag.


## `repositories.yaml`

```yaml
organization: <your organization>

team:
  - name: <a team name with admin privileges>
    level: admin
  - name: <a team name with read-only permissions>
    level: pull
  - name: <a-team-with-write-permissions>
    level: push

repositories:
  - my-repo
  - another-one
```

## Running

```
go run main.go
```

