# vlad

A very lightweight application deployer (VLAD), this is modeled on the idea of using a declaritive YAML based format to automate the deployment of applications. It leans on the standard templating build into Go, along with the [Amazon Web Services](https://aws.amazon.com/) (AWS) to enable deployment of [cloudformation](https://aws.amazon.com/cloudformation/) stacks, and other AWS resources.

# overview

The goals of this project are:

* Build a lightweight, fast deployment tool
* Enable tight integration with cloud services for configuration storage and encryption
* Focus on creating, configuring and deleting up cloud services

The things I am not planning to do with this project:

* Rebuild ansible
* Make a massive nebulas everything tool

Currently this tool is focused on the [AWS](http://aws.amazon.com/) cloud platform.

# usage

Given a simple YAML runbook file from the examples folder.

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
