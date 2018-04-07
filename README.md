# vlad

A very lightweight application deployer (VLAD), this is modeled on the idea of using a declaritive YAML based format to automate the deployment of applications. It leans on the standard templating build into Go, along with the [Amazon Web Services](https://aws.amazon.com/) (AWS) to enable deployment of [cloudformation](https://aws.amazon.com/cloudformation/) stacks, and other AWS resources.

# usage

Given a simple YAML file.

```yaml
vars:
    env: dev
    env_no: 1

tasks:
    - name: launch a lambda cloudformation stack
      cloudformation:
        stack_name: "deploy-lambda-{.env}-{.env_no}"
        template: "cloudformation/lambda.yml"
        disable_rollback: true
        tags:
            Owner: "wolfeidau"
```
