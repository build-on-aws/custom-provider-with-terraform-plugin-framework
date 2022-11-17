## Custom Provider with Terraform Plugin Framework

This repository contains a complete implementation of a custom provider built using the [Terraform plugin framework](https://developer.hashicorp.com/terraform/plugin/framework). It is used to teach, educate, and show the internals of a provider built with the latest SDK from [HashiCorp](https://www.hashicorp.com). Even if you are not looking to learn how to build custom providers, you may dial your troubleshooting skills to an expert level if you learn how one works behind the scenes. Plus, this provider is lots of fun to play with. The provider is called `buildonaws` and it allows you to maintain characters from comic books such as heros, super-heros, and villains.

### Requirements

* [Docker](https://www.docker.com/get-started)
* [Golang 1.18+](https://go.dev/dl)
* [Terraform 1.1.5+](https://www.terraform.io/downloads)

## ⚙️ Building the provider

The first thing you need to do to use this connector is to build it.

1. Install the following dependencies:

- [Golang 1.18+](https://go.dev/dl)

2. Execute the build file of the provider.

```bash
make install
```

💡 A file named `~/.terraform.d/plugins/aws.amazon.com/terraform/buildonaws/1.0/${OS_ARCH}/terraform-provider-buildonaws` will be created. This is your custom provider.

## ⬆️ Starting the provider backend

The provider uses [OpenSearch](https://opensearch.org) as the backend service to store and search for characters. Therefore, before playing with the provider, you first need to get OpenSearch up-and-running. For simplicity, this repository contains a [docker-compose.yml](./docker-compose.yml) file that you can use to execute OpenSearch as a container.

1. Install the following dependencies:

- [Docker](https://www.docker.com/get-started)

2. Start the containers using Docker Compose.

```bash
docker compose up -d
```

Wait until the container `opensearch` is started and healthy.

## ⏯ Playing with the provider

1. Install the following dependencies:

- [Terraform 1.1.5+](https://www.terraform.io/downloads)

2. Enter the `examples` directory.

```bash
cd examples
```

3. Initialize the provider plugins.

```bash
terraform init
```

4. Check the execution plan.

```bash
terraform plan
```

Once the command completes, you should see the following:

```bash
Terraform will perform the following actions:

  buildonaws_character.daredevil will be created
  + resource "buildonaws_character" "daredevil" {
      + fullname     = "Daredevil"
      + id           = (known after apply)
      + identity     = "Matt Murdock"
      + knownas      = "The man without fear"
      + last_updated = (known after apply)
      + type         = "super-hero"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + daredevil_secret_identity = "The secret identity of Daredevil is 'Matt Murdock'"
  + deadpool_is_knownas       = " is also known as ''"
╷
│ Warning: Datasource was not loaded
│ 
│   with data.buildonaws_character.deadpool,
│   on example.tf line 13, in data "buildonaws_character" "deadpool":
│   13: data "buildonaws_character" "deadpool" {
│ 
│ Reason: no character with the identity 'Wade Wilson'.
```

Note the warning in the end. This means that the data-source from the provider was not able to find any characters stored in OpenSearch whose identity is `Wade Wilson`. To create this character in the backend, execute the following command:

```bash
sh deadpool.sh
```

Check again:

```bash
terraform plan
```

Once the command completes, you should no more warnings.

```bash
Terraform will perform the following actions:

  buildonaws_character.daredevil will be created
  + resource "buildonaws_character" "daredevil" {
      + fullname     = "Daredevil"
      + id           = (known after apply)
      + identity     = "Matt Murdock"
      + knownas      = "The man without fear"
      + last_updated = (known after apply)
      + type         = "super-hero"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + daredevil_secret_identity = "The secret identity of Daredevil is 'Matt Murdock'"
  + deadpool_is_knownas       = "Deadpool is also known as 'Merch with a mouth'"
```

5. Apply the configuration

```bash
terraform apply -auto-approve
```

Once the command completes, you should see the outputs:

```bash
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

daredevil_secret_identity = "The secret identity of Daredevil is 'Matt Murdock'"
deadpool_is_knownas = "Deadpool is also known as 'Merch with a mouth'"
```

Once you are done playing with the provider, you can destroy the resources created using:

```bash
terraform destroy -auto-approve
```

## 🪲 Debugging the provider

This is actually an optional step, but if you wish to debug the connector code to learn its behavior by watching the code executing line by line, you can do so by using [delve](https://github.com/go-delve/delve).

1. Create a file named `.vscode/launch.json` with the following content:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Provider",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "env": {},
            "args": [
                "-debug",
            ]
        }
    ]
}
```

2. Set one or multiple breakpoints throughout the code.
3. Launch a new debugging session in Visual Studio Code.
4. In the debug console output you will see the following:

```bash
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:

	TF_REATTACH_PROVIDERS='{"aws.amazon.com/terraform/buildonaws":{"Protocol":"grpc","ProtocolVersion":6,"Pid":00000,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/jp/8jhflhbx6fj9_br8dwlfxn780000gr/T/plugin1062046873"}}}'

```

💡 Please note that your `TF_REATTACH_PROVIDERS` environment variable will differ from what is shown here. Please don't copy the example above. You need to copy the one generated in your machine.

5. Attach the Terraform CLI to the debugger

```bash
export TF_REATTACH_PROVIDERS='{"aws.amazon.com/terraform/buildonaws":{"Protocol":"grpc","ProtocolVersion":6,"Pid":00000,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/jp/8jhflhbx6fj9_br8dwlfxn780000gr/T/plugin1062046873"}}}'
```

6. Run any Terraform commands (plan, apply, destroy) that may trigger the breakpoints.

## ⬇️ Stopping the provider backend

1. Stop the containers using Docker Compose.

```bash
docker compose down
```

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the MIT-0 License. See the [LICENSE](./LICENSE) file.